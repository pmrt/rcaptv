package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/tracker"
)

func waitSig() os.Signal {
	sigint := make(chan os.Signal, 1)
	signal.Notify(
		sigint,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	return <-sigint
}

func main() {
	l := logger.New("tracker", "main")
	l.Info().Msg("Tracker starting")
	if !cfg.IsProd {
		l.Warn().Msg("[!] Running tracker in dev mode")
	}
	if cfg.Debug {
		l.Warn().Msg("log level: debug")
	} else {
		l.Info().Msg("log level: info")
	}

	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)

	sto := database.NewWithLogger(postgres.New(
		&database.StorageOptions{
			StorageHost:     cfg.PostgresHost,
			StoragePort:     cfg.PostgresPort,
			StorageUser:     cfg.PostgresUser,
			StoragePassword: cfg.PostgresPassword,
			StorageDbName:   cfg.PostgresDBName,

			StorageMaxIdleConns:    cfg.PostgresMaxIdleConns,
			StorageMaxOpenConns:    cfg.PostgresMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.PostgresConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.PostgresMigVersion,
			MigrationPath:    cfg.PostgresMigPath,
		}), l)
	hx := helix.NewWithLogger(&helix.HelixOpts{
		Creds: helix.ClientCreds{
			ClientID:     cfg.HelixClientID,
			ClientSecret: cfg.HelixClientSecret,
		},
		APIUrl:           cfg.APIUrl,
		EventsubEndpoint: cfg.EventSubEndpoint,
	}, l)

	go func() {
		if err := tracker.New(&tracker.TrackerOpts{
			Helix:   hx,
			Context: ctx,
			Storage: sto,
		}).Run(); err != nil {
			l.Panic().Err(err).Msg("tracker returned an error")
		}
	}()
	sig := waitSig()
	l.Warn().Msgf("Termination signal received [%s]. Attempting gracefully shutdown...", sig)
	l.Info().Msg("Tracker stopped")
	ctxCancel()
}

func init() {
	cfg.Setup()
}
