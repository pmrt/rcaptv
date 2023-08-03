package webserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/template/handlebars/v2"
	"github.com/rs/zerolog/log"

	"pedro.to/rcaptv/auth"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
)

type WebServer struct {
	sv       *fiber.App
	passport *auth.Passport
}

var scopes = []string{"user:read:email"}

// Starts the web server. Shutdown() must be handled.
func (sv *WebServer) StartAndListen(port string) error {
	l := log.With().Str("ctx", "webserver").Logger()

	l.Info().Msg("initializing webserver...")
	sv.passport.Start()
	app := sv.newServer()
	if err := app.Listen(":" + port); err != nil {
		return err
	}
	return nil
}

// Shutdown stops API services and server
func (sv *WebServer) Shutdown() error {
	l := log.With().Str("ctx", "webserver").Logger()
	l.Info().Msg("shutting down webserver...")
	defer sv.passport.Stop()
	return sv.sv.Shutdown()
}

func (sv *WebServer) newServer() *fiber.App {
	l := log.With().Str("ctx", "webserver").Logger()

	engine := handlebars.New(cfg.WebserverViewsDir, ".hbs")
	app := fiber.New(fiber.Config{
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    20 * time.Second,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		BodyLimit:       4 * 1024 * 1024,
		Concurrency:     256 * 1024,
		Views:           engine,
	})
	if cfg.IsProd {
		// in-memory ratelimiter for production. Use redis if needed in the future
		l.Info().Msgf("websv: ratelimit set to hits:%d, exp:%ds", cfg.WebserverRateLimitMaxConns, cfg.WebserverRateLimitExpSeconds)
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.IP() == "127.0.0.1"
			},
			Max:               cfg.WebserverRateLimitMaxConns,
			Expiration:        time.Duration(cfg.WebserverRateLimitMaxConns) * time.Second,
			LimiterMiddleware: limiter.SlidingWindow{},
			LimitReached: func(c *fiber.Ctx) error {
				l.Warn().Msgf("ratelimit reached (%s)", c.IP())
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}))
	}

	l.Info().Msg("websv: setting up encrypted cookies middleware")
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key:    cfg.CookieSecret,
		Except: []string{"csrf_", cookie.UserCookie},
	}))

	origins := strings.Join(cfg.Origins(), ", ")
	l.Info().Msgf("websv: setting up cors (domains: %s)", origins)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET, POST, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Accept-Language",
		AllowCredentials: true,
	}))

	l.Info().Msg("websv: setting up request handlers")
	app.Get(cfg.HealthEndpoint, func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).Send([]byte("ok"))
	})
	app.Get(cfg.LoginEndpoint, sv.passport.WithAuth, sv.passport.Login)
	app.Get(cfg.LogoutEndpoint, sv.passport.WithAuth, sv.passport.Logout)
	auth := app.Group(cfg.AuthEndpoint)
	auth.Get(cfg.AuthRedirectEndpoint, sv.passport.Callback)
	l.Info().Msgf("websv health: %s", cfg.HealthEndpoint)
	l.Info().Msgf("websv Login: %s", cfg.LoginEndpoint)
	l.Info().Msgf("websv Callback: %s", cfg.AuthEndpoint+cfg.AuthRedirectEndpoint)

	// react app
	app.Static("/", cfg.WebserverStaticDir)
	// catch all
	app.Static("*", cfg.WebserverIndexDir)
	sv.sv = app
	return app
}

func New(p *auth.Passport) *WebServer {
	sv := &WebServer{
		passport: p,
	}
	return sv
}
