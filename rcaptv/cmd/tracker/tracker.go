package main

import (
	"context"
	"time"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/tracker"
	"pedro.to/rcaptv/utils"

	"github.com/rs/zerolog/log"
)

func main() {
	l := log.With().Str("ctx", "main").Logger()
	l.Info().Msgf("tracker starting (v%s)", cfg.Version)
	if !cfg.IsProd {
		l.Warn().Msg("[!] running tracker in dev mode")
	}
	if cfg.Debug {
		l.Warn().Msg("log level: debug")
	} else {
		l.Info().Msg("log level: info")
	}

	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)

	l.Info().Msg("initializing database (postgres)")
	sto := database.New(postgres.New(
		&database.StorageOptions{
			StorageHost:     cfg.PostgresHost,
			StoragePort:     cfg.PostgresPort,
			StorageUser:     cfg.TrackerPostgresUser,
			StoragePassword: cfg.TrackerPostgresPassword,
			StorageDbName:   cfg.PostgresDBName,

			StorageMaxIdleConns:    cfg.PostgresMaxIdleConns,
			StorageMaxOpenConns:    cfg.PostgresMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.PostgresConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.PostgresMigVersion,
			MigrationPath:    cfg.PostgresMigPath,
		}))

	l.Info().Msg("initializing helix client (using credentials)")
	hx := helix.New(&helix.HelixOpts{
		Creds: helix.ClientCreds{
			ClientID:     cfg.HelixClientID,
			ClientSecret: cfg.HelixClientSecret,
		},
		APIUrl:           cfg.TwitchAPIUrl,
		EventsubEndpoint: cfg.EventSubEndpoint,
	})

	go func() {
		l.Info().Msg("starting tracker service")
		if err := tracker.New(&tracker.TrackerOpts{
			Helix:   hx,
			Context: ctx,
			Storage: sto,
		}).Run(); err != nil {
			l.Panic().Err(err).Msg("tracker returned an error")
		}
	}()
	sig := utils.WaitInterrupt()

	l.Info().Msgf("termination signal received [%s]. Attempting gracefully shutdown...", sig)
	l.Info().Msg("closing database")
	if err := sto.Stop(); err != nil {
		l.Warn().Err(err).Msg("error closing database")
	}
	l.Info().Msg("stopping tracker")
	ctxCancel()
}

func init() {
	cfg.Setup()
}
