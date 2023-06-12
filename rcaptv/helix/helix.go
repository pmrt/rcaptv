package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

const EstimatedSubscriptionJSONSize = 350

// Note: thinking about moving this to opts? Maybe it's time for a custom http
// client
const HttpMaxAttempts = 3
const HttpRetryDelay = time.Second * 5

var (
	ErrTooManyRequestAttempts = errors.New("no attempts left for performing requests")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrUnexpectedStatusCode   = errors.New("unexpected status code")
)

type ClientCreds struct {
	ClientID, ClientSecret string
}

type HelixOpts struct {
	creds ClientCreds

	APIUrl           string
	EventsubEndpoint string

	// Event handlers
	handleStreamOnline  func(evt *EventStreamOnline)
	handleStreamOffline func(evt *EventStreamOffline)

	// Webhook handlers
	handleRevocation func(evt *WebhookRevokePayload)
}

type Helix struct {
	ctx  context.Context
	opts *HelixOpts
	c    *http.Client
}

func (hx *Helix) CreateEventsubSubscription(sub *Subscription) error {
	b := struct {
		Type      string     `json:"type"`
		Version   string     `json:"version"`
		Condition *Condition `json:"condition"`
		Transport *Transport `json:"transport"`
	}{
		Type:      sub.Type,
		Version:   sub.Version,
		Condition: sub.Condition,
		Transport: sub.Transport,
	}

	buf := bytes.NewBuffer(make([]byte, 0, EstimatedSubscriptionJSONSize))
	if err := json.NewEncoder(buf).Encode(b); err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		hx.opts.APIUrl+hx.opts.EventsubEndpoint+"/subscriptions",
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := hx.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Expected 200 response, got " + fmt.Sprint(resp.StatusCode))
	}
	return nil
}

// Do handles some errors for resiliency and retries if possible. If
// Ratelimit-reset header is present during TooManyRequests errors it will
// retry after the reset time
//
// HttpMaxAttempts controls the number of attempts to retry.
// HttpRetryDelay controls the delay before a retry for 5XX errors.
//
// Do not use the body of the response as it will already be processed.
func (hx *Helix) Do(req *http.Request) (*http.Response, error) {
	return hx.doAtMost(req, HttpMaxAttempts)
}

func (hx *Helix) doAtMost(req *http.Request, attempts int) (*http.Response, error) {
	if attempts <= 0 {
		return nil, ErrTooManyRequestAttempts
	}

	if hx.opts.creds.ClientSecret != "" {
		req.Header.Set("Client-Id", hx.opts.creds.ClientID)
	}
	resp, err := hx.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Note: the body read is not limited since we trust the server responses. If
	// this changes in the future this is a good place for a limited reader.

	respondedAt, err := parseRespDate(resp.Header.Get("Date"))
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusAccepted:
		break
	case http.StatusTooManyRequests:
		d, err := untilRatelimitReset(
			resp.Header.Get("Ratelimit-Reset"),
			respondedAt,
		)
		if err != nil {
			return nil, err
		}
		time.Sleep(d + time.Second)
		attempts--
		return hx.doAtMost(req, attempts)
	case http.StatusInternalServerError:
	case http.StatusBadGateway:
	case http.StatusServiceUnavailable:
	case http.StatusGatewayTimeout:
		// Retry some 5XX twitch responses for resiliency
		time.Sleep(HttpRetryDelay)
		attempts--
		return hx.doAtMost(req, attempts)
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, ErrUnexpectedStatusCode
	}
	return resp, nil
}

func (hx *Helix) HandleStreamOnline(cb func(evt *EventStreamOnline)) {
	hx.opts.handleStreamOnline = cb
}

func (hx *Helix) HandleStreamOffline(cb func(evt *EventStreamOffline)) {
	hx.opts.handleStreamOffline = cb
}

func (hx *Helix) HandleRevocation(cb func(evt *WebhookRevokePayload)) {
	hx.opts.handleRevocation = cb
}

func parseRespDate(date string) (time.Time, error) {
	return time.Parse(time.RFC1123, date)
}

// untilRatelimitReset takes a reset unix timestamp and request timestamp and returns
// the time.Duration until the next reset
func untilRatelimitReset(reset string, respondedAt time.Time) (time.Duration, error) {
	ts64, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return time.Duration(0), err
	}
	resetAt := time.Unix(ts64, 0)
	if err != nil {
		return time.Duration(0), err
	}
	return resetAt.Sub(respondedAt), nil
}

// Exchange uses the client credentials to get a new http client with the
// corresponding token source, refreshing the token when needed. This http
// client injects the required Authorization header to the requests and will be
// used by the following requests.
//
// Must be used before using authenticated endpoints.
func (hx *Helix) Exchange() {
	o2 := &clientcredentials.Config{
		ClientID:     hx.opts.creds.ClientID,
		ClientSecret: hx.opts.creds.ClientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	hx.c = o2.Client(hx.ctx)
}

// NewWithoutExchange instantiates a new Helix client but without exchanging
// credentials for a token source. Useful for testing.
//
// Use New() if your helix client will use authenticated endpoints.
func NewWithoutExchange(opts *HelixOpts) *Helix {
	hx := &Helix{
		opts: opts,
		c:    http.DefaultClient,
		ctx:  context.Background(),
	}
	if hx.opts.handleStreamOnline == nil {
		hx.opts.handleStreamOnline = func(evt *EventStreamOnline) {}
	}
	if hx.opts.handleStreamOffline == nil {
		hx.opts.handleStreamOffline = func(evt *EventStreamOffline) {}
	}
	if hx.opts.handleRevocation == nil {
		hx.opts.handleRevocation = func(evt *WebhookRevokePayload) {}
	}
	return hx
}

func New(opts *HelixOpts) *Helix {
	hx := NewWithoutExchange(opts)
	hx.Exchange()
	return hx
}
