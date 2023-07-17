package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type VideoPeriod int

const (
	None VideoPeriod = iota
	All
	Day
	Week
	Month
)

func (p VideoPeriod) String() string {
	return [...]string{"none", "all", "day", "week", "month"}[p]
}

type VODParams struct {
	BroadcasterID  string
	GameID         string
	Lang           string
	Period         VideoPeriod
	After          string
	First          int
	StopAtVODID    string
	OnlyMostRecent bool

	Context context.Context
}

type VOD struct {
	VideoID        string    `json:"id" sql:"primary_key" alias:"vods.video_id"`
	BroadcasterID  string    `json:"user_id" alias:"vods.bc_id"`
	StreamID       string    `json:"stream_id" alias:"vods.stream_id"`
	CreatedAt      time.Time `json:"created_at" alias:"vods.created_at"`
	PublishedAt    time.Time `json:"published_at" alias:"vods.published_at"`
	DurationString string    `json:"duration,omitempty"`
	Lang           string    `json:"language" alias:"vods.lang"`
	Title          string    `json:"title" alias:"vods.title"`
	ThumbnailURL   string    `json:"thumbnail_url" alias:"vods.thumbnail_url"`
	ViewCount      int       `json:"view_count" alias:"vods.view_count"`

	Duration int32 `json:"duration_seconds" alias:"vods.duration_seconds"`
}

func (v *VOD) DurationSeconds() (int32, error) {
	if v.Duration != 0 {
		return v.Duration, nil
	}
	d, err := time.ParseDuration(v.DurationString)
	s := int32(d.Seconds())
	v.Duration = s
	return s, err
}

func ParseVODDurations(vods []*VOD) error {
	for _, vod := range vods {
		if _, err := vod.DurationSeconds(); err != nil {
			return err
		}
	}
	return nil
}

// Vods return the last videos of type VOD for a given broadcaster if specified
// in the parameters in order of time (last or more recent VODs first).
//
// p.StopAtVODID is used to stop asking for more clips after a given VODID.
//
// If p.StopAtVODID is empty, it will fetch and return only the most recent
// VOD.
func (hx *Helix) Vods(p *VODParams) ([]*VOD, error) {
	params := url.Values{}

	if p.BroadcasterID != "" {
		params.Add("user_id", p.BroadcasterID)
	}
	if p.GameID != "" {
		params.Add("game_id", p.GameID)
	}
	if p.Lang != "" {
		params.Add("language", p.Lang)
	}
	if p.Period != 0 {
		params.Add("period", p.Period.String())
	}
	if p.After != "" {
		params.Add("after", p.After)
	}
	if p.First == 0 {
		p.First = 100
	}

	pagFunc := func(vod *VOD, _ []*VOD) bool {
		return p.StopAtVODID == vod.VideoID
	}
	if p.OnlyMostRecent {
		p.First = 1
		pagFunc = func(vod *VOD, _ []*VOD) bool {
			return true
		}
	}
	params.Add("first", strconv.Itoa(p.First))
	params.Add("type", "archive")

	if p.Context == nil {
		p.Context = context.Background()
	}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/videos?%s", hx.APIUrl(), params.Encode()),
		nil,
	)
	req = req.WithContext(p.Context)
	if err != nil {
		return nil, err
	}
	vods, err := DoWithPagination[*VOD](hx, req, pagFunc, func(v *VOD) string {
		return v.VideoID
	})
	if err != nil {
		return nil, err
	}
	return vods, ParseVODDurations(vods)
}
