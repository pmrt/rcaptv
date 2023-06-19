package helix

import (
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
}

type VOD struct {
	VideoID       string    `json:"id"`
	BroadcasterID string    `json:"user_id"`
	StreamID      string    `json:"stream_id"`
	CreatedAt     time.Time `json:"created_at"`
	PublishedAt   time.Time `json:"published_at"`
	Duration      string    `json:"duration"`
	Lang          string    `json:"language"`
	Title         string    `json:"title"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	ViewCount     int       `json:"view_count"`

	durationSeconds *int
}

func (v *VOD) DurationSeconds() (int, error) {
	if v.durationSeconds != nil {
		return *v.durationSeconds, nil
	}
	d, err := time.ParseDuration(v.Duration)
	s := int(d.Seconds())
	v.durationSeconds = &s
	return s, err
}

// Vods return the last videos of type VOD for a given broadcaster if specified
// in the parameters in order of time (last or more recent VODs first).
//
// p.StopAtVODID is used to stop asking for more clips after a given VODID.
//
// If p.StopAtVODID is empty, it will fetch and return only the most recent
// VOD.
func (hx *Helix) Vods(p *VODParams) ([]VOD, error) {
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

	pagFunc := func(vod VOD, _ []VOD) bool {
		return p.StopAtVODID == vod.VideoID
	}
	if p.OnlyMostRecent {
		p.First = 1
		pagFunc = func(vod VOD, _ []VOD) bool {
			return true
		}
	}
	params.Add("first", strconv.Itoa(p.First))
	params.Add("type", "archive")
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/videos?%s", hx.APIUrl(), params.Encode()),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return DoWithPagination[VOD](hx, req, pagFunc)
}
