package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"pedro.to/rcaptv/auth"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/utils"
)

// TODO: split vods and clips into their files
//

var ErrDbLocalClips = errors.New("Unexpected error while retrieving local clips")

type APIOpts struct {
	Storage database.Storage

	// Max difference between start/end for clips.
	ClipsMaxPeriodDiffHours int

	ClientID, ClientSecret string
	HelixAPIUrl            string
	HelixEventsubEndpoint  string
}

type API struct {
	db *sql.DB
	hx *helix.Helix

	clipsMaxPeriodDiffHours int

	sv       *fiber.App
	passport *auth.Passport
}

type ResultsMode string

const (
	ModeLocal  ResultsMode = "local"
	ModeHybrid ResultsMode = "hybrid"
	ModeRemote ResultsMode = "remote"
)

type APIResponse[T any] struct {
	Data   T           `json:"data"`
	Errors []string    `json:"errors"`
	Mode   ResultsMode `json:"mode"`
}

func NewResponse[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{
		Data:   data,
		Errors: make([]string, 0, 2),
	}
}

type VodsResponse struct {
	Vods []*helix.VOD `json:"vods"`
}

func (a *API) Vods(c *fiber.Ctx) error {
	resp := NewResponse(&VodsResponse{
		Vods: make([]*helix.VOD, 0, 5),
	})
	resp.Mode = ModeLocal

	username := c.Query("username")
	vid := c.Query("vid")
	after := c.Query("after")

	vids := make([]string, 0, 1)
	if username == "" {
		if vid == "" {
			if after == "" {
				resp.Errors = append(resp.Errors, "Missing username, after or vid")
				return c.Status(http.StatusBadRequest).JSON(resp)
			}
		} else {
			vids = append(vids, vid)
		}
	}
	ext, err := strconv.Atoi(c.Query("extend", "0"))
	if err != nil {
		resp.Errors = append(resp.Errors, "Bad extend value")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	vods, err := repo.Vods(a.db, &repo.VodsParams{
		VideoIDs:   vids,
		BcUsername: username,
		Extend:     utils.Min(ext, 5),
		After:      after,
		Context:    c.Context(),
	})
	if err != nil {
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.Status(http.StatusInternalServerError).JSON(resp)
	}
	if len(vods) == 0 {
		if username != "" {
			// TODO - this is a good thing to log, so we know which channels are our users interested in
			resp.Errors = append(resp.Errors, fmt.Sprintf("Username '%s' not found. The channel may not be tracked by us.", username))
		} else if vid != "" {
			resp.Errors = append(resp.Errors, fmt.Sprintf("VOD '%s' not found", vid))
		} else if after != "" {
			resp.Errors = append(resp.Errors, fmt.Sprintf("No VODS after '%s' found", after))
		}
		return c.Status(http.StatusNotFound).JSON(resp)
	}

	resp.Data.Vods = append(resp.Data.Vods, vods...)
	return c.Status(http.StatusOK).JSON(resp)
}

type ClipsResponse struct {
	Clips []*helix.Clip `json:"clips"`
}

// Clips
// - `bid` string Broacaster ID
// - `started_at` string Start range time of creation of the clip in RFC3339
// - `ended_at` string End range time of creation of the clip in RFC3339
//
// If user is logged in it will presented with results from twitch api and
// local tracked clips, deduplicated and merged. When merged we get the most
// updated view_count value and vod_offset, video_id if available
//
// If user is not logged in only local tracked clips will be returned. View
// count may be outdated. Local clips won't return dangling clips (with
// vod_offset=null or empty video_id) by default.
//
// Note: Twitch API does not provide a way to fetch clips by video_id. An
// alternative is to ask for bid+start+end of the stream. This may leave out
// some interesting clips created after the stream and include clips from other
// vods (unless filtered by video_id). To mitigate this we use an hybrid
// solution when user is logged in we fetch clips from twitch api, we retrieve
// clips from database, deduplicate and merge them. On the frontend, clips
// should be filtered by vod_offset=null and the corresponding video_id.
func (a *API) Clips(c *fiber.Ctx) error {
	if auth.IsLoggedIn(c) {
		return a.hybridClips(c)
	}
	return a.localClips(c)
}

func (a *API) hybridClips(c *fiber.Ctx) error {
	resp := NewResponse(new(ClipsResponse))
	resp.Data.Clips = make([]*helix.Clip, 0, 200)
	resp.Mode = ModeHybrid

	params, errs := a.getClipParams(c)
	if len(errs) > 0 {
		resp.Errors = errs
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	res := make(chan []*helix.Clip, 2)
	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	// clips from twitch api
	g.Go(func() error {
		clips, err := a.hx.DeepClips(&helix.DeepClipsParams{
			ClipsParams: &helix.ClipsParams{
				BroadcasterID:            params.bid,
				StartedAt:                params.started,
				EndedAt:                  params.ended,
				StopViewsThreshold:       cfg.ClipViewThreshold,
				ViewsThresholdWindowSize: cfg.ClipViewWindowSize,
				SkipDeduplication:        true,
				Context:                  ctx,
			},
			MaxDeepLvl: cfg.ClipTrackingMaxDeepLevel,
		})
		if err != nil {
			return err
		}
		res <- clips
		return nil
	})
	// clips from db, previously tracked with our tracker
	g.Go(func() error {
		localClips, err := repo.Clips(a.db, &repo.ClipsParams{
			BroadcasterID:   params.bid,
			StartedAt:       params.started,
			EndedAt:         params.ended,
			ExcludeDangling: true,
			Context:         ctx,
		})
		if err != nil {
			return ErrDbLocalClips
		}
		res <- localClips
		return nil
	})
	if err := a.checkErr(c, g.Wait()); err != nil {
		if errors.Is(err, helix.ErrItemsEmpty) {
			resp.Errors = append(resp.Errors,
				fmt.Sprintf("No clips found for the provided streamer (bid:'%s'). Are clips enabled for this streamer?",
					params.bid),
			)
			return c.JSON(resp)
		}
		if errors.Is(err, ErrDbLocalClips) {
			resp.Errors = append(resp.Errors, err.Error())
			return c.JSON(resp)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			resp.Errors = append(resp.Errors, "Timeout when fetching clips in hybrid mode. Try again.")
			c.Status(http.StatusGatewayTimeout)
			return c.JSON(resp)
		}
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.JSON(resp)
	}
	for i := 0; i < 2; i++ {
		clips := <-res
		resp.Data.Clips = append(resp.Data.Clips, clips...)
	}
	resp.Data.Clips = helix.Deduplicate(resp.Data.Clips, func(c *helix.Clip) string {
		return c.ClipID
	}, func(a *helix.Clip, b *helix.Clip) *helix.Clip {
		// preserve videoID if present in a or b
		a.VideoID = utils.CoalesceString(a.VideoID, b.VideoID)
		// preserve vodOffset if present in a or b
		a.VODOffsetSeconds = utils.Coalesce(a.VODOffsetSeconds, b.VODOffsetSeconds)
		// get most updated value of viewCount
		a.ViewCount = utils.Max(a.ViewCount, b.ViewCount)
		return a
	})
	return c.Status(http.StatusOK).JSON(resp)
}

func (a *API) localClips(c *fiber.Ctx) error {
	resp := NewResponse(new(ClipsResponse))
	resp.Data.Clips = make([]*helix.Clip, 0, 1)
	resp.Mode = ModeLocal

	params, errs := a.getClipParams(c)
	if len(errs) > 0 {
		resp.Errors = errs
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	localClips, err := repo.Clips(a.db, &repo.ClipsParams{
		BroadcasterID:   params.bid,
		StartedAt:       params.started,
		EndedAt:         params.ended,
		ExcludeDangling: true,
		Context:         c.Context(),
	})
	if err != nil {
		resp.Errors = append(resp.Errors, "Unexpected error while retrieving local clips")
		return c.JSON(resp)
	}
	if len(localClips) == 0 {
		resp.Errors = append(resp.Errors,
			fmt.Sprintf("No clips found for the provided streamer (bid:'%s'). Check if the streamer has clips enabled. If it does, try logging in using the 'login with Twitch' button for the hybrid mode, which allows us to perform requests directly to the Twitch API with more flexible rate limits.",
				params.bid),
		)
		return c.Status(http.StatusNotFound).JSON(resp)
	}
	resp.Data.Clips = localClips
	return c.Status(http.StatusOK).JSON(resp)
}

// checkErr checks errors for helix and session state
func (a *API) checkErr(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, helix.ErrUnauthorized) {
		auth.ClearAuthCookies(c)
		c.Status(http.StatusUnauthorized)
		return helix.ErrUnauthorized
	}

	if errors.Is(err, helix.ErrItemsEmpty) {
		c.Status(http.StatusNotFound)
		return helix.ErrItemsEmpty
	}

	c.Status(http.StatusInternalServerError)
	return fmt.Errorf("unexpected error: %w", err)
}

type clipParams struct {
	bid     string
	started time.Time
	ended   time.Time
}

func (a *API) getClipParams(c *fiber.Ctx) (clipParams, []string) {
	errors := make([]string, 0, 4)
	bid := c.Query("bid")
	if bid == "" {
		errors = append(errors, "Missing bid")
	}
	startedAt := c.Query("started_at")
	if startedAt == "" {
		errors = append(errors, "Missing started_at")
	}
	endedAt := c.Query("ended_at")
	if endedAt == "" {
		errors = append(errors, "Missing ended_at")
	}
	started, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		errors = append(errors, "Invalid 'started_at'")
	}
	ended, err := time.Parse(time.RFC3339, endedAt)
	if err != nil {
		errors = append(errors, "Invalid 'ended_at'")
	}
	if ended.Sub(started) > time.Duration(a.clipsMaxPeriodDiffHours)*time.Hour {
		errors = append(errors, "period between 'started_at' and 'ended_at' is too large")
	}
	return clipParams{
		bid:     bid,
		started: started,
		ended:   ended,
	}, errors
}

// Starts the api server. Shutdown() must be handled.
func (a *API) StartAndListen(port string) error {
	l := log.With().Str("ctx", "apiserver").Logger()

	l.Info().Msg("initializing apiserver...")
	a.passport.Start()
	app := a.newServer()
	if err := app.Listen(":" + port); err != nil {
		return err
	}
	return nil
}

func (a *API) Shutdown() error {
	l := log.With().Str("ctx", "apiserver").Logger()
	l.Info().Msg("shutting down apiserver...")
	defer a.passport.Stop()
	return a.sv.Shutdown()
}

func (a *API) newServer() *fiber.App {
	l := log.With().Str("ctx", "apiserver").Logger()

	app := fiber.New(fiber.Config{
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    20 * time.Second,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		BodyLimit:       4 * 1024 * 1024,
		Concurrency:     256 * 1024,
		ProxyHeader:     fiber.HeaderXForwardedFor,
	})
	if cfg.IsProd {
		// in-memory ratelimiter for production. Use redis if needed in the future
		l.Info().Msgf("apisv: ratelimit set to hits:%d, exp:%ds", cfg.APIRateLimitMaxConns, cfg.APIRateLimitExpSeconds)
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.IP() == "127.0.0.1"
			},
			Max:               cfg.APIRateLimitMaxConns,
			Expiration:        time.Duration(cfg.APIRateLimitMaxConns) * time.Second,
			LimiterMiddleware: limiter.SlidingWindow{},
			LimitReached: func(c *fiber.Ctx) error {
				l.Warn().Msgf("ratelimit reached (%s)", c.IP())
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}))
	} else {
		l.Info().Msg("apisv: setting up request logger middleware")
		app.Use(logger.Fiber())
	}
	l.Info().Msg("apisv: setting up encrypted cookies middleware")
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key:    cfg.CookieSecret,
		Except: []string{"csrf_", cookie.UserCookie},
	}))

	origins := strings.Join(cfg.Origins(), ", ")
	l.Info().Msgf("apisv: setting up cors (domains: %s)", origins)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET, POST, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))

	l.Info().Msg("apisv: setting up request handlers")
	app.Get(cfg.HealthEndpoint, func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).Send([]byte("ok"))
	})
	v1 := app.Group(cfg.APIEndpoint)
	v1.Get(cfg.APIVodsEndpoint, a.Vods)

	hx := v1.Group(cfg.APIHelixEndpoint, a.passport.WithAuth)
	hx.Get(cfg.APIValidateEndpoint, a.passport.ValidateSession)
	hx.Get(cfg.APIClipsEndpoint, a.Clips)

	l.Info().Msgf("apisv health: %s", cfg.HealthEndpoint)
	l.Info().Msgf("apisv vods: %s", cfg.APIEndpoint+cfg.APIVodsEndpoint)
	l.Info().Msgf("apisv validate: %s", cfg.APIEndpoint+cfg.APIHelixEndpoint+cfg.APIValidateEndpoint)
	l.Info().Msgf("apisv clips: %s", cfg.APIEndpoint+cfg.APIHelixEndpoint+cfg.APIClipsEndpoint)
	a.sv = app
	return app
}

func New(auth *auth.Passport, opts APIOpts) *API {
	if opts.ClipsMaxPeriodDiffHours == 0 {
		opts.ClipsMaxPeriodDiffHours = 24 * 7
	}
	db := opts.Storage.Conn()
	api := &API{
		passport: auth,
		db:       db,
		hx: helix.NewWithUserTokens(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     opts.ClientID,
				ClientSecret: opts.ClientSecret,
			},
			APIUrl: opts.HelixAPIUrl,
		}),
		clipsMaxPeriodDiffHours: opts.ClipsMaxPeriodDiffHours,
	}
	return api
}
