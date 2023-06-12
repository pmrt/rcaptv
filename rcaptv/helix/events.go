package helix

import (
	"time"
)

// Twitch stream types
// See "Stream Online Event" https://dev.twitch.tv/docs/eventsub/eventsub-reference#events
const (
	StreamLive       string = "live"
	StreamPlaylist   string = "playlist"
	StreamWatchParty string = "watch_party"
	StreamPremiere   string = "premiere"
	StreamRerun      string = "rerun"
)

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	StartedAt time.Time `json:"started_at"`
}

type Broadcaster struct {
	ID       string `json:"broadcaster_user_id"`
	Login    string `json:"broadcaster_user_login"`
	Username string `json:"broadcaster_user_name"`
}

// Twitch Events
// See https://dev.twitch.tv/docs/eventsub/eventsub-reference#events

type EventStreamOnline struct {
	*Event
	*Broadcaster
}

type EventStreamOffline struct {
	*Broadcaster
}
