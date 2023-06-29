package main

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rs/zerolog/log"
	"pedro.to/rcaptv/api"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/utils"
)

func main() {
	l := log.With().Str("ctx", "main").Logger()
	l.Info().Msgf("starting rcaptv server (v%s)", cfg.Version)
	if !cfg.IsProd {
		l.Warn().Msg("[!] running rcaptv server in dev mode")
	}
	l.Info().Msgf("starting rcaptv server (v%s)", cfg.Version)

	sto := database.New(postgres.New(
		&database.StorageOptions{
			StorageHost:     cfg.PostgresHost,
			StoragePort:     cfg.PostgresPort,
			StorageUser:     cfg.RcaptvPostgresUser,
			StoragePassword: cfg.RcaptvPostgresPassword,
			StorageDbName:   cfg.PostgresDBName,

			StorageMaxIdleConns:    cfg.PostgresMaxIdleConns,
			StorageMaxOpenConns:    cfg.PostgresMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.PostgresConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.PostgresMigVersion,
			MigrationPath:    cfg.PostgresMigPath,
	}))

	app := fiber.New(fiber.Config{
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 20 * time.Second,
		ReadBufferSize: 4096,
		WriteBufferSize: 4096,
		BodyLimit: 4 * 1024 * 1024,
		Concurrency: 256*1024,
	})
	if cfg.IsProd {
		// in-memory ratelimiter for production. Use redis if needed in the future
		l.Info().Msgf("ratelimit set to hits:%d, exp:%ds", cfg.RateLimitMaxConns, cfg.RateLimitExpSeconds)
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.IP() == "127.0.0.1"
			},
			Max: cfg.RateLimitMaxConns,
			Expiration: time.Duration(cfg.RateLimitExpSeconds) * time.Second,
			LimiterMiddleware: limiter.SlidingWindow{},
			LimitReached: func(c *fiber.Ctx) error {
				l.Warn().Msgf("ratelimit reached (%s)", c.IP())
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}))
		// TODO - ssl
	} else {
		// allow cors if in dev mode
		app.Use(func(c *fiber.Ctx) error {
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type")
			return c.Next()
		})
	}
	app.Use(logger.Fiber())

	hx := helix.New(&helix.HelixOpts{
		Creds: helix.ClientCreds{
			ClientID:     cfg.HelixClientID,
			ClientSecret: cfg.HelixClientSecret,
		},
		APIUrl:           cfg.APIUrl,
		EventsubEndpoint: cfg.EventSubEndpoint,
	})
	api := api.New(api.APIOpts{
		Helix: hx,
		Storage: sto,
		ClipsMaxPeriodDiffHours: cfg.ClipsMaxPeriodDiffHours,
})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).Send([]byte("ok"))
	})
	v1 := app.Group("/v1")
	v1.Get("/vods", api.Vods)
	v1.Get("/clips", api.Clips)

	go func() {
		l.Info().Msgf("rcaptv server listening on port:%s", cfg.APIPort)
	 if err := app.Listen(":"+cfg.APIPort); err != nil {
		 l.Panic().Err(err).Msg("rcaptv server returned an error")
	 }
	}()
	sig := utils.WaitInterrupt()
	l.Info().Msgf("termination signal received [%s]. Attempting gracefully shutdown...", sig)
	l.Info().Msg("stopping rcaptv server")
	if err := app.Shutdown(); err != nil {
		l.Error().Err(err).Msg("rcaptv server returned an error while shutting down")
	}
}

func init() {
	cfg.Setup()
}