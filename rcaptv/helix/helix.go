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
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"

	"pedro.to/rcaptv/utils"
)

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
)

type CtxHelixKey string

const CtxHelixTokenSource CtxHelixKey = "helix_token_source"

const CtxHelixCustomQuery CtxHelixKey = "helix_custom_query"

type CustomQueryOpts struct {
	UseClientID bool
}

func ContextWithTokenSource(tk *oauth2.Token, opts NotifyReuseTokenSourceOpts) context.Context {
	src := NotifyReuseTokenSource(tk, opts)
	return context.WithValue(context.Background(), CtxHelixTokenSource, src)
}

func ContextWithCustomQueryOpts(opts *CustomQueryOpts) context.Context {
	return context.WithValue(context.Background(), CtxHelixCustomQuery, opts)
}

const EstimatedSubscriptionJSONSize = 350

// Note: thinking about moving this to opts? Maybe it's time for a custom http
// client
const (
	HttpMaxAttempts                     = 3
	HttpRetryDelay                      = time.Second * 5
	HttpMaxClientResponseReadLimitBytes = 1 * MB
)

var (
	ErrTooManyRequestAttempts = errors.New("no attempts left for performing requests")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrUnexpectedStatusCode   = errors.New("unexpected status code")
	ErrBodyResponseTooBig     = errors.New("response body too big")
	ErrBodyEmpty              = errors.New("response body empty")
	ErrItemsEmpty             = errors.New("no items returned")
	ErrBadRequest             = errors.New("bad request")
	ErrNotFound               = errors.New("not found")
	ErrInvalidContext         = errors.New("invalid context for current request")
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

type RFC3339Timestamp time.Time

func (ts *RFC3339Timestamp) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`)
	if value == "" {
		return nil
	}
	if value == "null" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	*ts = RFC3339Timestamp(t)
	return nil
}

func (ts *RFC3339Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*ts).Format(time.RFC3339) + `"`), nil
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
	Creds ClientCreds

	APIUrl           string
	EventsubEndpoint string
	ValidateEndpoint string

	// Event handlers
	HandleStreamOnline  func(evt *EventStreamOnline)
	HandleStreamOffline func(evt *EventStreamOffline)

	// Webhook handlers
	HandleRevocation func(evt *WebhookRevokePayload)
}

// Helix client for Twitch Helix API
//
// Helix is safe for concurrent access if opts are never mutated after
// initialization
type Helix struct {
	ctx              context.Context
	opts             *HelixOpts
	defaultClient    *http.Client
	defaultQueryOpts *CustomQueryOpts

	useUserTokens bool
}

// ClientID returns the client id which the helix client was initializated
// with.
//
// IMPORTANT: Do not use hx.opts.creds.ClientID directly and do not ever mutate
// it or the client would need mutexes for safe concurrent access.
func (hx *Helix) ClientID() string {
	return hx.opts.Creds.ClientID
}

// ClientSecret returns the client secret which the helix client was
// initializated with.
//
// IMPORTANT: Do not use hx.opts.creds.ClientSecret directly and do not ever
// mutate it or the client would need mutexes for safe concurrent access.
func (hx *Helix) ClientSecret() string {
	return hx.opts.Creds.ClientSecret
}

// APIUrl returns the twitch API url which the helix client was initializated
// with.
//
// IMPORTANT: Do not use hx.opts.APIUrl directly and do not ever mutate it or
// the client would need mutexes for safe concurrent access.
func (hx *Helix) APIUrl() string {
	return hx.opts.APIUrl
}

func (hx *Helix) ValidateEndpoint() string {
	return hx.opts.ValidateEndpoint
}

// EventsubEndpoint returns the twitch eventsub endpoint which the helix client
// was initialized with.
//
// IMPORTANT: Do not use hx.opts.EventsubEndpoint directly and do not ever
// mutate it or the client would need mutexes for safe concurrent access.
func (hx *Helix) EventsubEndpoint() string {
	return hx.opts.EventsubEndpoint
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
//
// If a deduplicateKeyFn is passed, it will deduplicate the results with the
// key returned by the function. If deduplicateKeyFn is nil, the results will
// be returned intact.
func DoWithPagination[T any](
	hx *Helix, req *http.Request,
	stopFunc func(item T, all []T) bool,
	deduplicateKeyFn func(i T) string,
) ([]T, error) {
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

		items := parsed.Data
		if len(items) == 0 {
			return nil, ErrItemsEmpty
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
	if deduplicateKeyFn != nil {
		return Deduplicate(all, deduplicateKeyFn), nil
	}
	return all, nil
}

func (hx *Helix) doAtMost(req *http.Request, attemptsLeft int) (*HttpResponse, error) {
	lctx := log.With().
		Str("ctx", "helix")

	if attemptsLeft <= 0 {
		return nil, ErrTooManyRequestAttempts
	}

	ctx := req.Context()

	qopts := hx.defaultQueryOpts
	if o, ok := ctx.Value(CtxHelixCustomQuery).(*CustomQueryOpts); ok && o != nil {
		qopts = o
	}

	if hx.ClientSecret() != "" && qopts.UseClientID {
		req.Header.Set("Client-Id", hx.ClientID())
	}

	attempts := utils.Abs(attemptsLeft-HttpMaxAttempts) + 1
	l := lctx.
		Int("attempts", attempts).
		Str("endpoint", req.URL.Path).
		Logger()

	c := hx.defaultClient
	if hx.useUserTokens {
		ts, ok := ctx.Value(CtxHelixTokenSource).(oauth2.TokenSource)
		if !ok {
			l.Err(ErrInvalidContext).Msg("invalid context")
			return nil, ErrInvalidContext
		}
		c = oauth2.NewClient(context.Background(), ts)
	}
	resp, err := c.Do(req)
	if err != nil {
		l.Info().Msgf("%s-> %s %s (attempts:%d) <- ('%s')",
			req.Method, req.URL.Path, req.URL.RawQuery, attempts, err.Error(),
		)
		return nil, err
	}
	l.Info().Msgf("%s-> %s %s (attempts:%d) <- %d",
		req.Method, req.URL.Path, req.URL.RawQuery, attempts, resp.StatusCode,
	)
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
		l.Warn().Msgf("ratelimit reached for %s: %s %s. Retry: %s (attempts:%d)",
			req.Method, req.URL.Path, req.URL.RawQuery, d, attempts,
		)
		time.Sleep(d + time.Second)
		attemptsLeft--
		return hx.doAtMost(req, attemptsLeft)
	case http.StatusInternalServerError:
	case http.StatusBadGateway:
	case http.StatusServiceUnavailable:
	case http.StatusGatewayTimeout:
		l.Warn().Msgf("got HTTP %d from provider for %s: %s %s. Retry %s (attempts:%d)",
			resp.StatusCode, req.Method, req.URL.Path, req.URL.RawQuery,
			HttpRetryDelay, attempts,
		)
		// Retry some 5XX twitch responses for resiliency
		time.Sleep(HttpRetryDelay)
		attemptsLeft--
		return hx.doAtMost(req, attemptsLeft)
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, ErrUnexpectedStatusCode
	}
	return &HttpResponse{
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

func (hx *Helix) HandleStreamOnline(cb func(evt *EventStreamOnline)) {
	hx.opts.HandleStreamOnline = cb
}

func (hx *Helix) HandleStreamOffline(cb func(evt *EventStreamOffline)) {
	hx.opts.HandleStreamOffline = cb
}

func (hx *Helix) HandleRevocation(cb func(evt *WebhookRevokePayload)) {
	hx.opts.HandleRevocation = cb
}

func parseRespDate(date string) (time.Time, error) {
	return time.Parse(http.TimeFormat, date)
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
		ClientID:     hx.ClientID(),
		ClientSecret: hx.ClientSecret(),
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	hx.defaultClient = o2.Client(hx.ctx)
}

// NewWithoutExchange instantiates a new Helix client but without exchanging
// credentials for a token source. Useful for testing or when using user tokens
// instead of app tokens (oauth tokens that are retrieved from cookies)
//
// Use New() if your helix client will use authenticated endpoints with app
// tokens and NewWithUserTokens() if your will use user tokens instead.
func NewWithoutExchange(opts *HelixOpts, c ...*http.Client) *Helix {
	hx := &Helix{
		opts:          opts,
		defaultClient: http.DefaultClient,
		defaultQueryOpts: &CustomQueryOpts{
			UseClientID: true,
		},
		ctx: context.Background(),
	}
	if len(c) == 1 {
		hx.defaultClient = c[0]
	}
	if hx.opts.HandleStreamOnline == nil {
		hx.opts.HandleStreamOnline = func(evt *EventStreamOnline) {}
	}
	if hx.opts.HandleStreamOffline == nil {
		hx.opts.HandleStreamOffline = func(evt *EventStreamOffline) {}
	}
	if hx.opts.HandleRevocation == nil {
		hx.opts.HandleRevocation = func(evt *WebhookRevokePayload) {}
	}
	if hx.opts.ValidateEndpoint == "" {
		hx.opts.ValidateEndpoint = TwitchValidateEndpoint
	}
	return hx
}

func Deduplicate[T any](s []T, keyFn func(i T) string) []T {
	r := make([]T, 0, len(s))
	ht := map[string]bool{}
	for _, item := range s {
		if k := keyFn(item); !ht[k] {
			r = append(r, item)
			ht[k] = true
		}
	}
	return r
}

func New(opts *HelixOpts) *Helix {
	hx := NewWithoutExchange(opts)
	hx.Exchange()
	return hx
}

func NewWithUserTokens(opts *HelixOpts) *Helix {
	hx := NewWithoutExchange(opts)
	hx.useUserTokens = true
	return hx
}
