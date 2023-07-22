package api

import (
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

	"pedro.to/rcaptv/auth"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/utils"
)

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

	sv   *fiber.App
	auth *auth.Passport
}

type APIResponse[T any] struct {
	Data   T        `json:"data"`
	Errors []string `json:"errors"`
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
// Note: Twitch API does not provide a way to fetch clips by video_id.
// Alternative is to ask for bid+start+end of the stream. This may leave out
// some interesting clips created after the stream and include clips from other
// vods. One potential solution is to merge with tracker clips in server and
// filter by vod id in the client
func (a *API) Clips(c *fiber.Ctx) error {
	resp := NewResponse(&ClipsResponse{
		Clips: make([]*helix.Clip, 0, 10),
	})

	bid := c.Query("bid")
	if bid == "" {
		resp.Errors = append(resp.Errors, "Missing bid")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	started_at := c.Query("started_at")
	if started_at == "" {
		resp.Errors = append(resp.Errors, "Missing started_at")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	ended_at := c.Query("ended_at")
	if ended_at == "" {
		resp.Errors = append(resp.Errors, "Missing ended_at")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	started, err := time.Parse(time.RFC3339, started_at)
	if err != nil {
		resp.Errors = append(resp.Errors, "Invalid 'started_at'")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	ended, err := time.Parse(time.RFC3339, ended_at)
	if err != nil {
		resp.Errors = append(resp.Errors, "Invalid 'ended_at'")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	if ended.Sub(started) > time.Duration(a.clipsMaxPeriodDiffHours)*time.Hour {
		resp.Errors = append(resp.Errors, "period between 'started_at' and 'ended_at' is too large")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	clips, err := a.hx.DeepClips(&helix.DeepClipsParams{
		ClipsParams: &helix.ClipsParams{
			BroadcasterID:            bid,
			StartedAt:                started,
			EndedAt:                  ended,
			StopViewsThreshold:       cfg.ClipViewThreshold,
			ViewsThresholdWindowSize: cfg.ClipViewWindowSize,
			Context:                  c.UserContext(),
		},
		MaxDeepLvl: cfg.ClipTrackingMaxDeepLevel,
	})
	c, err = a.checkErr(c, err)
	if err != nil {
		if errors.Is(err, helix.ErrItemsEmpty) {
			resp.Errors = append(resp.Errors, fmt.Sprintf("No clips found for the provided streamer (bid:'%s'). Are clips enabled for this streamer?", bid))
			return c.JSON(resp)
		}
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.JSON(resp)
	}
	resp.Data.Clips = append(resp.Data.Clips, clips...)
	return c.Status(http.StatusOK).JSON(resp)
}

func (a *API) checkErr(c *fiber.Ctx, err error) (*fiber.Ctx, error) {
	if err == nil {
		return c, nil
	}

	if errors.Is(err, helix.ErrUnauthorized) {
		auth.ClearAuthCookies(c)
		return c.Status(http.StatusUnauthorized), helix.ErrUnauthorized
	}

	if errors.Is(err, helix.ErrItemsEmpty) {
		return c.Status(http.StatusNotFound), helix.ErrItemsEmpty
	}

	return c.Status(http.StatusInternalServerError), fmt.Errorf("unexpected error: %w", err)
}

func (a *API) StartAndListen(port string) error {
	l := log.With().Str("ctx", "apiserver").Logger()

	l.Info().Msg("initializing apiserver...")
	app := a.newServer()
	if err := app.Listen(":" + port); err != nil {
		return err
	}
	return nil
}

func (a *API) Shutdown() error {
	l := log.With().Str("ctx", "apiserver").Logger()
	l.Info().Msg("shutting down apiserver...")
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

	hx := v1.Group(cfg.APIHelixEndpoint, a.auth.WithAuth)
	hx.Get(cfg.APIValidateEndpoint, a.auth.ValidateSession)
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
		auth: auth,
		db:   db,
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
