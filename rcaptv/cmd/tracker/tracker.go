package main

import (
	"time"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/logger"
)

func main() {
	l := logger.New("tracker", "main")
	l.Info().Msg("Tracker starting")
	if !cfg.IsProd {
		l.Warn().Msg("[!] Running tracker in dev mode")
	}

	database.NewWithLogger(postgres.New(
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
}

func init() {
	cfg.Setup()
}
