package helix

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
)

const EstimatedSubscriptionJSONSize = 350

// Note: thinking about moving this to opts? Maybe it's time for a custom http
// client
const HttpMaxAttempts = 3
const HttpRetryDelay = time.Second * 5
const HttpMaxClientResponseReadLimitBytes = 1 * MB

var (
	ErrTooManyRequestAttempts = errors.New("no attempts left for performing requests")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrUnexpectedStatusCode   = errors.New("unexpected status code")
	ErrBodyResponseTooBig     = errors.New("response body too big")
	ErrBodyEmpty              = errors.New("response body empty")
)

type HttpResponse struct {
	Body       []byte
	StatusCode int
}

type ClientCreds struct {
	ClientID, ClientSecret string
}

type Pagination struct {
	Cursor string
}

// modReqQuery provides a way to easily edit a particular parameter in a `req`
// request.
func modReqQuery(req *http.Request, key, value string) error {
	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	values.Set(key, value)
	req.URL.RawQuery = values.Encode()

	return nil
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

// Do handles some errors for resiliency and retries if possible. If
// Ratelimit-reset header is present during TooManyRequests errors it will
// retry after the reset time
//
// HttpMaxAttempts controls the number of attempts to retry.
// HttpRetryDelay controls the delay before a retry for 5XX errors.
//
// Do not use the body of the response as it will already be processed.
func (hx *Helix) Do(req *http.Request) (*HttpResponse, error) {
	return hx.doAtMost(req, HttpMaxAttempts)
}

type PaginationManyObj[T any] struct {
	Data       []T
	Pagination *Pagination
}

// Do handles a http request with twitch pagination.
//
// stopFunc(item, all) is called after reading and adding each item to the all
// slice. The boolean value returned by the stopFunc() defines when to stop
// performing more requests.
//
// If stopFunc() returns false and all the elements in the request are
// processed, DoWithPagination will perform a new request using the cursor from
// the previous one. If stopFunc() returns true while processing a element, the
// loop will break and no more requests will be performed.
func DoWithPagination[T any](hx *Helix, req *http.Request, stopFunc func(item T, all []T) bool) ([]T, error) {
	var (
		resp   *HttpResponse
		parsed *PaginationManyObj[T]
		err    error
		// Twitch page size is 100 items max
		all = make([]T, 0, 100)
	)
PaginationLoop:
	for {
		resp, err = hx.Do(req)
		if err != nil {
			return nil, err
		}

		if len(resp.Body) == 0 {
			return nil, ErrBodyEmpty
		}

		if err = json.Unmarshal(resp.Body, &parsed); err != nil {
			return nil, err
		}

		for _, item := range parsed.Data {
			item := item
			all = append(all, item)
			if stopFunc(item, all) {
				break PaginationLoop
			}
		}

		if parsed.Pagination == nil {
			break
		} else if parsed.Pagination.Cursor == "" {
			break
		}

		if err = modReqQuery(req, "after", parsed.Pagination.Cursor); err != nil {
			return nil, err
		}
		parsed = nil
	}
	return all, nil
}

func (hx *Helix) doAtMost(req *http.Request, attempts int) (*HttpResponse, error) {
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
	r := io.LimitReader(resp.Body, HttpMaxClientResponseReadLimitBytes)
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, ErrBodyResponseTooBig
	}

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
	return &HttpResponse{
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
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
