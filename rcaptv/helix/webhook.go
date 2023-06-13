package helix

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/utils"
)

var (
	WebhookEventHMACPrefix       = []byte("sha256=")
	WebhookEventHMACPrefixLength = len(WebhookEventHMACPrefix)
)

const (
	// Twitch webhook events
	// See https://dev.twitch.tv/docs/eventsub/handling-webhook-events
	WebhookEventNotification string = "notification"
	WebhookEventVerification string = "webhook_callback_verification"
	WebhookEventRevocation   string = "revocation"

	SubStreamOnline  string = "stream.online"
	SubStreamOffline string = "stream.offline"
)

const (
	// Twitch webhook headers
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events#list-of-request-headers
	WebhookHeaderID        = "Twitch-Eventsub-Message-Id"
	WebhookHeaderTimestamp = "Twitch-Eventsub-Message-Timestamp"
	WebhookHeaderSignature = "Twitch-Eventsub-Message-Signature"
	WebhookHeaderType      = "Twitch-Eventsub-Message-Type"
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

type WebhookHeaders struct {
	ID        string
	Timestamp string
	Signature string
	Type      string
	Body      []byte
}

func WebhookHeadersFromFiber(c *fiber.Ctx) *WebhookHeaders {
	return &WebhookHeaders{
		ID:        c.Get(WebhookHeaderID),
		Timestamp: c.Get(WebhookHeaderTimestamp),
		Signature: c.Get(WebhookHeaderSignature),
		Type:      c.Get(WebhookHeaderType),
		Body:      c.Body(),
	}
}

func (evt *WebhookHeaders) Valid(secret []byte) bool {
	// Important note: DO NOT mutate id, sig and ts, they are meant to be read-only
	var (
		id   = utils.StringToByte(evt.ID)
		ts   = utils.StringToByte(evt.Timestamp)
		sig  = utils.StringToByte(evt.Signature)
		body = evt.Body
	)

	mac := hmac.New(sha256.New, secret)
	mac.Write(id)
	mac.Write(ts)
	mac.Write(body)
	hash := mac.Sum(nil)
	l := len(hash)
	hexHash := make([]byte, hex.EncodedLen(l), hex.EncodedLen(l)+WebhookEventHMACPrefixLength)
	hex.Encode(hexHash, hash)
	hexHash = utils.Prepend(hexHash, WebhookEventHMACPrefix)
	return hmac.Equal(sig, hexHash)
}

type WebhookHandler struct {
	secret []byte
	hx     *Helix
	// For testing
	fakeNow time.Time
}

func (h *WebhookHandler) handler(c *fiber.Ctx) error {
	now := time.Now()
	if !cfg.IsProd {
		if !h.fakeNow.IsZero() {
			now = h.fakeNow
		}
	}

	headers := WebhookHeadersFromFiber(c)
	if !headers.Valid(h.secret) {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid signature")
	}

	// Mitigate replay attacks. Ignore events with valid signature and ts older
	// than 10min from now. Note: we are not storing the event IDs.
	t, err := time.Parse(time.RFC3339, headers.Timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid timestamp")
	}
	if now.Add(-10 * time.Minute).After(t) {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired timestamp")
	}

	switch headers.Type {
	case WebhookEventNotification:
		var resp *WebhookNotificationPayload
		if err := c.BodyParser(&resp); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
		}
		return h.handleEvent(resp)
	case WebhookEventVerification:
		var resp *WebhookVerificationPayload
		if err := c.BodyParser(&resp); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
		}
		if resp.Challenge == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Empty challenge")
		}
		return c.SendString(resp.Challenge)
	case WebhookEventRevocation:
		var resp *WebhookRevokePayload
		if err := c.BodyParser(&resp); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
		}
		go h.hx.opts.handleRevocation(resp)
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Unknown Twitch-Eventsub-Message-Type header")
	}
	return nil
}

func (h *WebhookHandler) handleEvent(resp *WebhookNotificationPayload) error {
	// handlers may involve long-running tasks. Twitch expects a quick response
	// from the webhook server, so we carry out these tasks in a different
	// goroutine
	switch resp.Subscription.Type {
	case SubStreamOnline:
		go h.hx.opts.handleStreamOnline(&EventStreamOnline{
			Event: &Event{
				ID:        resp.Event.Event.ID,
				Type:      resp.Event.Type,
				StartedAt: resp.Event.StartedAt,
			},
			Broadcaster: &Broadcaster{
				ID:       resp.Event.Broadcaster.ID,
				Login:    resp.Event.Broadcaster.Login,
				Username: resp.Event.Broadcaster.Username,
			},
		})
	case SubStreamOffline:
		go h.hx.opts.handleStreamOffline(&EventStreamOffline{
			Broadcaster: &Broadcaster{
				ID:       resp.Event.Broadcaster.ID,
				Login:    resp.Event.Broadcaster.Login,
				Username: resp.Event.Broadcaster.Username,
			},
		})
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Unknown notification subscription type")
	}

	return nil
}
