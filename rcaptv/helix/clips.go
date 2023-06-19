package helix

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

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
}

type Clip struct {
	ClipID           string  `json:"id" sql:"primary_key" alias:"clips.clip_id"`
	BroadCasterID    string  `json:"broadcaster_id" alias:"clips.bc_id"`
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
	Clips []Clip
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

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/clips?%s", hx.APIUrl(), params.Encode()),
		nil,
	)
	if err != nil {
		return nil, err
	}

	if p.ViewsThresholdWindowSize == 0 {
		p.ViewsThresholdWindowSize = 4
	}
	if p.StopViewsThreshold == 0 {
		p.StopViewsThreshold = 8
	}

	bop := bufop.New(p.ViewsThresholdWindowSize)
	t := float32(p.StopViewsThreshold)
	stopped := false
	clips, err := DoWithPagination[Clip](hx, req, func(item Clip, all []Clip) bool {
		bop.PutInt(item.ViewCount)
		if bop.Avg() < t {
			stopped = true
			return true
		}
		return false
	})
	return &ClipResponse{
		Clips:      clips,
		IsComplete: stopped,
	}, err
}
