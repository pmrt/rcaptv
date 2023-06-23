package tracker

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/rs/zerolog/log"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

var (
	ErrEmptyVODs  = errors.New("no VODs found")
	ErrEmptyClips = errors.New("no clips found")
)

type lastVODTable map[string]string

func (t lastVODTable) FromDB(db *sql.DB) error {
	rows, err := repo.LastVODByStreamer(db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		t[row.BroadcasterID] = row.VodID
	}
	return nil
}

func NewLastVODTable(estStreamers int) lastVODTable {
	return make(lastVODTable, estStreamers)
}

type Tracker struct {
	db                *sql.DB
	ctx               context.Context
	hx                *helix.Helix
	lastVIDByStreamer lastVODTable

	TrackingCycleMinutes     int
	ClipTrackingMaxDeepLevel int
	ClipTrackingWindowHours  int
	ClipViewThreshold        int
	ClipViewWindowSize       int

	// Useful for testing. Run won't FetchVods/Clips if true. Not available in
	// production mode
	FakeRun bool
	stopped bool
}

func (t *Tracker) Run() error {
	l := log.With().Str("ctx", "tracker").Logger()

	l.Info().Msg("fetching streamer list from database")
	streamers, err := repo.Tracked(t.db)
	if err != nil {
		return err
	}
	lenbc := len(streamers)
	l.Info().Msgf("%d streamers loaded", lenbc)

	l.Info().Msg("loading last VOD table")
	t.lastVIDByStreamer = NewLastVODTable(lenbc)
	t.lastVIDByStreamer.FromDB(t.db)

	l.Info().Msg("initializing scheduler")
	bs := newBalancedSchedule(BalancedScheduleOpts{
		CycleSize:          uint(t.TrackingCycleMinutes),
		EstimatedStreamers: uint(lenbc),
	})
	for _, streamer := range streamers {
		bs.Add(streamer.BcID)
	}
	cs := bs.opts.CycleSize
	l.Info().
		Msgf("starting scheduler real-time tracking (cycle_size=%d, estimated_streamers=%d)",
			cs, bs.opts.EstimatedStreamers,
		)
	bs.Start()

	for {
		select {
		// For every scheduler tick we get the minute (or unit we're using) and the
		// list of streamers to be invoked within that minute
		case m := <-bs.RealTime():
			for _, bid := range m.Streamers {
				bid := bid
				if !cfg.IsProd && t.FakeRun {
					l.Warn().Msg("skipping run in FakeRun mode")
					continue
				}

				// Execute sequantially streamer requests. Streamer by streamer. We are
				// in no hurry and we don't want to be rate-limited.
				//
				// As we are adding more and more streamers and we get closer to the
				// limit and if rate-limiting is limiting by seconds too, we may want
				// to change the unit of minutes to seconds to make it easier to crunch
				// the numbers to find out rate limits and the right cycle size
				clips, err := t.FetchClips(bid)
				if err != nil {
					if errors.Is(err, ErrEmptyClips) {
						// TODO - track maybe track this too. Add no clips count to db
						l.Warn().Msgf("no clips found (bid:%s)", bid)
					} else {
						l.Err(err).Msgf("failed to fetch clips (bid:%s)", bid)
					}
				}
				vods, err := t.FetchVods(bid)
				if err != nil {
					if errors.Is(err, ErrEmptyVODs) {
						// TODO - update tracked_channels.seen_inactive_count
						l.Warn().Msgf("no VODs found (bid:%s)", bid)
					} else {
						l.Err(err).Msg("failed to fetch VODs")
					}
				}

				lenc, lenv := len(clips), len(vods)
				if lenc > 0 {
					if err := repo.UpsertClips(t.db, clips); err != nil {
						l.Err(err).Msgf("failed to upsert clips (clips:%d)",
							lenc,
						)
					}
				}
				if lenv > 0 {
					if err := repo.UpsertVods(t.db, vods); err != nil {
						l.Err(err).Msgf("failed to upsert VODs (VODs:%d)",
							lenv,
						)
					}
				}
				l.Info().Msgf(
					"[balanced_key:%d/%d] updated clips:%d and VODs:%d (bid:%s)",
					m.Min, cs-1, lenc, lenv, bid,
				)
			}
		case <-t.ctx.Done():
			l.Info().Msg("stopping scheduler real-time tracking")
			t.stopped = true
			bs.Cancel()
			return t.ctx.Err()
		}
	}
}

// FetchVods retrieves VODS for a given broadcaster ID up to the last vod ID,
// including the last VOD ID in the result. Then it updates the lastVODs table
// with the new most recent VOD. The last VOD ID is included and fetched again
// just in case it was updated since the last time (e.g.: the duration)
//
// If last VOD = "" FetchVods() will fetch only the most recent VOD and update
// the table with it.
func (t *Tracker) FetchVods(bid string) ([]*helix.VOD, error) {
	lastvid := t.lastVIDByStreamer[bid]

	opts := &helix.VODParams{
		BroadcasterID: bid,
		Period:        helix.Week,
		StopAtVODID:   lastvid,
	}
	if lastvid == "" {
		opts.OnlyMostRecent = true
	}

	vods, err := t.hx.Vods(opts)
	if err != nil {
		if errors.Is(err, helix.ErrItemsEmpty) {
			return nil, ErrEmptyVODs
		}
		return nil, err
	}
	t.lastVIDByStreamer[bid] = vods[0].VideoID
	return vods, nil
}

// FetchClips for a given broadcaster ID in a rolling window specified by
// ClipTrackingWindowHours. Clips obtained will stop if they don't meet a view
// threshold in a rolling window average. Both specified correspondingly by
// ClipViewThreshold and ClipViewWindowSize
//
// Twitch API returns after consuming all the pages in pagination a maximum of
// 1000 items for a given request. To avoid incomplete data FetchClips performs
// a deep fetch, where it first performs a request for our predefined window.
// If the response didn't reach the minimum threshold because the 1000 clips
// have a large number of views, we consider the response incomplete and we
// would divide the window in 2, performing another 2 requests, etc. up to a
// maximum specified by ClipTrackingMaxDeepLevel.
//
// Maximum number of requests per streamer is `2^(maxlevel-1)+2(maxlevel-2) ...
// 2^(maxlevel-m) ... 2^0` e.g.: if max_level=3;max_requests/streamer=7
//
// For example, if ClipTrackingWindowHours is set to 168 hours (7 days), and
// the response is marked as incomplete, we would perform another 2 requests
// for the ranges 0-84 and 84-168 hours in the corresponding rolling window
func (t *Tracker) FetchClips(bid string) ([]*helix.Clip, error) {
	now := time.Now()
	from := now.Add(-time.Duration(t.ClipTrackingWindowHours) * time.Hour)
	return t.deepFetchClips(bid, 1, from, now)
}

func (t *Tracker) deepFetchClips(bid string, lvl int, from time.Time, to time.Time) ([]*helix.Clip, error) {
	l := log.With().Str("ctx", "tracker").Logger()
	clipsResp, err := t.hx.Clips(&helix.ClipsParams{
		BroadcasterID:            bid,
		StopViewsThreshold:       t.ClipViewThreshold,
		ViewsThresholdWindowSize: t.ClipViewWindowSize,
		StartedAt:                from,
		EndedAt:                  to,
	})
	if err != nil {
		if errors.Is(err, helix.ErrItemsEmpty) {
			return nil, ErrEmptyClips
		}
		return nil, err
	}
	if clipsResp.IsComplete {
		return clipsResp.Clips, nil
	}
	// If next level is too deep, we stop here and return the current results
	if lvl+1 > t.ClipTrackingMaxDeepLevel {
		l.Warn().Msgf("incomplete clip results after clip_tracking_max_deep_level=%d "+
			"reached for period from=%s to=%s (bid:%s) ",
			cfg.ClipTrackingMaxDeepLevel, from.Format(time.RFC3339), to.Format(time.RFC3339), bid)
		return clipsResp.Clips, nil
	}

	nReqs := math.Pow(2, float64(lvl))
	partHours := float64(t.ClipTrackingWindowHours) / nReqs
	all := make([]*helix.Clip, 0, 100*2)
	l.Debug().Msgf("(bid:%s) incomplete clip results for period from=%s to=%s. "+
		"Deepening (lvl:%d/%d, part_hours:%f, n_reqs:%f)",
		bid, from.Format(time.RFC3339), to.Format(time.RFC3339), lvl,
		t.ClipTrackingMaxDeepLevel, partHours, nReqs,
	)
	// left and right in the binary tree
	for i := 0; i < 2; i++ {
		to := from.Add(time.Duration(partHours) * time.Hour)
		clips, err := t.deepFetchClips(
			bid,
			lvl+1,
			from,
			to,
		)
		if err != nil {
			return nil, err
		}
		all = append(all, clips...)
		from = to
	}
	return all, nil
}

type TrackerOpts struct {
	Context context.Context
	Storage database.Storage
	Helix   *helix.Helix

	TrackingCycleMinutes     int
	ClipTrackingMaxDeepLevel int
	ClipTrackingWindowHours  int
	ClipViewThreshold        int
	ClipViewWindowSize       int
}

func New(opts *TrackerOpts) *Tracker {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.TrackingCycleMinutes == 0 {
		opts.TrackingCycleMinutes = cfg.TrackingCycleMinutes
	}
	if opts.ClipTrackingMaxDeepLevel == 0 {
		opts.ClipTrackingMaxDeepLevel = cfg.ClipTrackingMaxDeepLevel
	}
	if opts.ClipTrackingWindowHours == 0 {
		opts.ClipTrackingWindowHours = cfg.ClipTrackingWindowHours
	}
	if opts.ClipViewThreshold == 0 {
		opts.ClipViewThreshold = cfg.ClipViewThreshold
	}
	if opts.ClipViewWindowSize == 0 {
		opts.ClipViewWindowSize = cfg.ClipViewWindowSize
	}

	tk := &Tracker{
		ctx:                      opts.Context,
		hx:                       opts.Helix,
		TrackingCycleMinutes:     opts.TrackingCycleMinutes,
		ClipTrackingMaxDeepLevel: opts.ClipTrackingMaxDeepLevel,
		ClipTrackingWindowHours:  opts.ClipTrackingWindowHours,
		ClipViewThreshold:        opts.ClipViewThreshold,
		ClipViewWindowSize:       opts.ClipViewWindowSize,
	}
	if opts.Storage != nil {
		tk.db = opts.Storage.Conn()
	}
	return tk
}
