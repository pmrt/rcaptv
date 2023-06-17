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
	ClipID           string  `json:"id"`
	BroadCasterID    string  `json:"broadcaster_id"`
	VideoID          string  `json:"video_id"`
	CreatedAt        string  `json:"created_at"`
	CreatorID        string  `json:"creator_id"`
	CreatorName      string  `json:"creator_name"`
	Title            string  `json:"title"`
	GameID           string  `json:"game_id"`
	Lang             string  `json:"language"`
	ThumbnailURL     string  `json:"thumbnail_url"`
	DurationSeconds  float32 `json:"duration"`
	ViewCount        int     `json:"view_count"`
	VODOffsetSeconds *int    `json:"vod_offset"`
}

func (hx *Helix) Clips(p *ClipsParams) ([]Clip, error) {
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
	return DoWithPagination[Clip](hx, req, func(item Clip, all []Clip) bool {
		bop.PutInt(item.ViewCount)
		return bop.Avg() < t
	})
}
