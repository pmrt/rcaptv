package helix

import "time"

const (
	SubStreamOnline  string = "stream.online"
	SubStreamOffline string = "stream.offline"
)

type WebhookNotificationPayload struct {
	Subscription *Subscription `json:"subscription"`
	Event        struct {
		*Event
		*Broadcaster
	} `json:"event"`
}

type WebhookVerificationPayload struct {
	Challenge    string        `json:"challenge"`
	Subscription *Subscription `json:"subscription"`
}

type WebhookRevokePayload struct {
	Subscription *Subscription `json:"subscription"`
}

type Subscription struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Type      string     `json:"type"`
	Version   string     `json:"version"`
	Cost      int        `json:"cost"`
	Condition *Condition `json:"condition"`
	Transport *Transport `json:"transport"`
	CreatedAt time.Time  `json:"created_at"`
}

type Condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id"`
}

type Transport struct {
	Method   string `json:"method"`
	Callback string `json:"callback"`
	Secret   string `json:"secret"`
}
