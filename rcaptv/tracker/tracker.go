package tracker

import (
	"context"
	"database/sql"
	"sync"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/repo"
)

type Tracker struct {
	db  *sql.DB
	ctx context.Context
	hx  *helix.Helix
}

func (t *Tracker) Run() error {
	l := logger.New("tracker", "tracker")

	l.Info().Msg("=> starting tracker service")
	l.Info().
		Msg("=> => fetching streamer list from database")
	streamers, err := repo.Tracked(t.db)
	if err != nil {
		return err
	}
	len := len(streamers)
	l.Info().Msgf("=> => => %d streamers loaded", len)

	l.Info().
		Msg("=> => initializing scheduler")
	bs := newBalancedSchedule(BalancedScheduleOpts{
		CycleSize:          uint(cfg.TrackingCycleMinutes),
		EstimatedStreamers: uint(len),
	})
	for _, streamer := range streamers {
		bs.Add(streamer.BcID)
	}
	l.Info().
		Uint("EstimatedStreamers", bs.opts.EstimatedStreamers).
		Uint("CycleSize", bs.opts.CycleSize).
		Msg("starting scheduler real-time tracking")
	bs.Start()

	for {
		select {
		// For every scheduler tick we get the minute (or unit we're using) and the
		// list of streamers to be invoked within that minute
		case m := <-bs.RealTime():
			var wg sync.WaitGroup
			for _, bid := range m.Streamers {
				bid := bid
				l.Debug().
					Str("bid", bid).
					Msg("fetching streamer clips and vods")

				// Execute sequantially streamer requests. Streamer by streamer. We are
				// in no hurry and we don't want to be rate-limited.
				//
				// As we are adding more and more streamers and we get closer to the
				// limit and if rate-limiting is limiting by seconds too, we may want
				// to change the unit of minutes to seconds to make it easier to crunch
				// the numbers to find out rate limits and the right cycle size
				wg.Add(1)
				go func() {
					// get clips
					// get vods
					wg.Done()
				}()
				wg.Wait()
			}
		case <-t.ctx.Done():
			l.Info().Msg("stopping scheduler real-time tracking")
			bs.Cancel()
			return t.ctx.Err()
		}
	}
}

type TrackerOpts struct {
	Context context.Context
	Storage database.Storage
	Helix   *helix.Helix
}

func New(opts *TrackerOpts) *Tracker {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	return &Tracker{
		ctx: opts.Context,
		db:  opts.Storage.Conn(),
		hx:  opts.Helix,
	}
}
