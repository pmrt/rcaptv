package main

import (
	"time"

	"github.com/rs/zerolog/log"

	"pedro.to/rcaptv/api"
	"pedro.to/rcaptv/auth"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/utils"
)

func main() {
	l := log.With().Str("ctx", "api").Logger()
	l.Info().Msgf("starting api server (v%s)", cfg.Version)
	if !cfg.IsProd {
		l.Warn().Msg("[!] running api server in dev mode")
	}

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
	passport := auth.New(auth.PassportOps{
		Storage:      sto,
		ClientID:     cfg.HelixClientID,
		ClientSecret: cfg.HelixClientSecret,
		HelixAPIUrl:  cfg.TwitchAPIUrl,
	})
	apisv := api.New(passport, api.APIOpts{
		Storage:                 sto,
		ClipsMaxPeriodDiffHours: cfg.ClipsMaxPeriodDiffHours,
		ClientID:                cfg.HelixClientID,
		ClientSecret:            cfg.HelixClientSecret,
		HelixAPIUrl:             cfg.TwitchAPIUrl,
	})
	go func() {
		if err := apisv.StartAndListen(cfg.APIPort); err != nil {
			l.Panic().Err(err).Msg("")
		}
	}()
	sig := utils.WaitInterrupt()
	l.Info().Msgf("termination signal received [%s]. Attempting to gracefully shutdown...", sig)
	l.Info().Msg("stopping api server")
	if err := apisv.Shutdown(); err != nil {
		l.Panic().Err(err).Msg("")
	}
}

func init() {
	cfg.Setup()
}
