package helix

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"pedro.to/rcaptv/bufop"
)

type ClipsParams struct {
	BroadcasterID string
	GameID        string
	StartedAt     time.Time
	EndedAt       time.Time
	First         int
	After         string
	// Views threshold to stop fetching for more clips. The threshold is the
	// minimum average views of the clips in the window to be returned. When the
	// average in the rolling window is less than the threshold when checking a
	// given clip, no more clips will be fetched and the current []Clip will be
	// returned
	StopViewsThreshold int
	// Number of clips used to determine the threshold in a rolling window
	ViewsThresholdWindowSize int

	SkipDeduplication bool
	Context           context.Context
}

type Clip struct {
	ClipID           string  `json:"id" sql:"primary_key" alias:"clips.clip_id"`
	BroadcasterID    string  `json:"broadcaster_id" alias:"clips.bc_id"`
	VideoID          string  `json:"video_id" alias:"clips.video_id"`
	CreatedAt        string  `json:"created_at" alias:"clips.created_at"`
	CreatorID        string  `json:"creator_id" alias:"clips.creator_id"`
	CreatorName      string  `json:"creator_name" alias:"clips.creator_name"`
	Title            string  `json:"title" alias:"clips.title"`
	GameID           string  `json:"game_id" alias:"clips.game_id"`
	Lang             string  `json:"language" alias:"clips.lang"`
	ThumbnailURL     string  `json:"thumbnail_url" alias:"clips.thumbnail_url"`
	DurationSeconds  float32 `json:"duration" alias:"clips.duration_seconds"`
	ViewCount        int     `json:"view_count" alias:"clips.view_count"`
	VODOffsetSeconds *int    `json:"vod_offset" alias:"clips.vod_offset"`
}

type ClipResponse struct {
	Clips []*Clip
	// Twitch only returns up to 1000 items. IsComplete is false if after
	// requesting throughout the entire pagination, the view threshold was never
	// triggered, which is indicative that there could be more clips that meet
	// the view threshold
	IsComplete bool
}

func (hx *Helix) Clips(p *ClipsParams) (*ClipResponse, error) {
	params := url.Values{}

	if p.BroadcasterID != "" {
		params.Add("broadcaster_id", p.BroadcasterID)
	}
	if p.GameID != "" {
		params.Add("game_id", p.GameID)
	}
	if !p.StartedAt.IsZero() {
		params.Add("started_at", p.StartedAt.Format(time.RFC3339))
	}
	if !p.EndedAt.IsZero() {
		params.Add("ended_at", p.EndedAt.Format(time.RFC3339))
	}
	if p.After != "" {
		params.Add("after", p.After)
	}
	if p.First == 0 {
		p.First = 100
	}
	params.Add("first", strconv.Itoa(p.First))

	if p.Context == nil {
		p.Context = context.Background()
	}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/clips?%s", hx.APIUrl(), params.Encode()),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(p.Context)

	if p.ViewsThresholdWindowSize == 0 {
		p.ViewsThresholdWindowSize = 4
	}
	if p.StopViewsThreshold == 0 {
		p.StopViewsThreshold = 8
	}

	var dedupFn func(c *Clip) string
	if !p.SkipDeduplication {
		dedupFn = func(c *Clip) string {
			return c.ClipID
		}
	}

	bop := bufop.New(p.ViewsThresholdWindowSize)
	t := float32(p.StopViewsThreshold)
	stopped := false
	clips, err := DoWithPagination[*Clip](hx, req, func(item *Clip, all []*Clip) bool {
		bop.PutInt(item.ViewCount)
		if bop.Avg() < t {
			stopped = true
			return true
		}
		return false
	}, dedupFn)
	return &ClipResponse{
		Clips:      clips,
		IsComplete: stopped,
	}, err
}

type DeepClipsParams struct {
	*ClipsParams
	MaxDeepLvl int

	windowHours float64
}

// DeepClips is similar to Clips but it will try to fetch all the clips if the
// end of pagination is reached, up to a given MaxDeepLvl.
//
// If we request twitch clips in a given period and we want to get all the
// clips up to a given minimum views threshold, we may get to the final page of
// the pagination and we haven't yet reached the min. views threshold. This is
// because twitch only returns a maximum of 1000 elements for the entire
// pagination. This could be the case especially if the min. viewsthreshold is
// very low and the start/end period very long (which gives opportunity for the
// streamer to have more clips with more views).
//
// DeepClips tackles this problem by first attempting to fetch the clips up to
// the specified threshold and period range, just like Clips() would do. Then
// if the stop function was never invoked, that is: we went through all the
// clips and the min. threshold was never met, we consider the request
// incomplete and we divide the period (or time window) in half and try again
// and we go on recursively until every response is considered complete or the
// MaxDeepLvl is reached. The result is the sum of all the completed responses
// corresponding to every period part. The result periods should be exclusive
// but because the dynamic nature of the API and the pagination, some clips may
// be duplicated between Clips() results even if we deduplicate within Clips().
// So we deduplicate again the final result.
//
// For example, if the window is initially set to 168 hours (7 days), and the
// response is marked as incomplete, we would perform another 2 requests for
// the ranges 0-84 and 84-168 hours in the corresponding time window
func (hx *Helix) DeepClips(p *DeepClipsParams) ([]*Clip, error) {
	p.windowHours = p.EndedAt.Sub(p.StartedAt).Hours()
	skipDedup := p.SkipDeduplication
	// skip dedup anyway for every hx.Clips, since we'll perform our own
	// deduplication afterwards.
	p.SkipDeduplication = true

	if skipDedup {
		// if p.SkipDeduplication=true was explicity passed down, skip our
		// deduplication too
		return hx.deepFetchClips(*p, 1, p.StartedAt, p.EndedAt)
	}
	clips, err := hx.deepFetchClips(*p, 1, p.StartedAt, p.EndedAt)
	return Deduplicate(clips, func(c *Clip) string {
		return c.ClipID
	}), err
}

func (hx *Helix) deepFetchClips(p DeepClipsParams, lvl int, from time.Time, to time.Time) ([]*Clip, error) {
	l := log.With().Str("ctx", "helix").Logger()
	clipsResp, err := hx.Clips(&ClipsParams{
		BroadcasterID:            p.BroadcasterID,
		StopViewsThreshold:       p.StopViewsThreshold,
		ViewsThresholdWindowSize: p.ViewsThresholdWindowSize,
		Context:                  p.Context,
		SkipDeduplication:        p.SkipDeduplication,
		StartedAt:                from,
		EndedAt:                  to,
	})
	if err != nil {
		return nil, err
	}
	if clipsResp.IsComplete {
		return clipsResp.Clips, nil
	}
	// If next level is too deep, we stop here and return the current results
	if lvl+1 > p.MaxDeepLvl {
		l.Warn().Msgf("incomplete clip results after clip_tracking_max_deep_level=%d "+
			"reached for period from=%s to=%s (bid:%s) ",
			p.MaxDeepLvl, from.Format(time.RFC3339), to.Format(time.RFC3339), p.BroadcasterID)
		return clipsResp.Clips, nil
	}

	nReqs := math.Pow(2, float64(lvl))
	partHours := float64(p.windowHours) / nReqs
	all := make([]*Clip, 0, 100*2)
	l.Debug().Msgf("(bid:%s) incomplete clip results for period from=%s to=%s. "+
		"Deepening (lvl:%d/%d, part_hours:%f, n_reqs:%f)",
		p.BroadcasterID, from.Format(time.RFC3339), to.Format(time.RFC3339), lvl,
		p.MaxDeepLvl, partHours, nReqs,
	)
	// left and right in the binary tree
	for i := 0; i < 2; i++ {
		to := from.Add(time.Duration(partHours) * time.Hour)
		clips, err := hx.deepFetchClips(
			p,
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
