package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/utils"
)

type Storage interface {
	Ping(ctx context.Context) error
	Migrate() error
	Conn() *sql.DB
	Opts() *StorageOptions
	Stop() error
}

type StorageOptions struct {
	StorageHost     string
	StoragePort     string
	StorageUser     string
	StoragePassword string
	StorageDbName   string

	StorageMaxIdleConns    int
	StorageMaxOpenConns    int
	StorageConnMaxLifetime time.Duration
	StorageConnTimeout     time.Duration

	MigrationVersion int
	MigrationPath    string
	DebugMode        bool
}

func New(sto Storage) Storage {
	opts := sto.Opts()
	lctx := log.With().
		Str("ctx", "database")
	if !cfg.IsProd {
		lctx.
			Str("host", opts.StorageHost).
			Str("port", opts.StoragePort).
			Str("db", opts.StorageDbName).
			Str("user", opts.StorageUser).
			Str("pass", utils.TruncateSecret(opts.StoragePassword, 3))
	}
	l := lctx.Logger()

	l.Info().Msg("pinging database")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		opts.StorageConnTimeout,
	)
	defer cancel()
	if err := sto.Ping(ctx); err != nil {
		l.Panic().Err(err).Msg("")
	}
	l.Info().Msg("connection successful")

	if cfg.SkipMigrations {
		l.Info().Msg("skipping migrations")
		return sto
	}

	l.Info().Msgf("attempting to apply migrations (v%d @ %s)", opts.MigrationVersion, opts.MigrationPath)
	if err := sto.Migrate(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			l.Info().Msg("no changes were made")
		} else {
			l.Panic().Err(err).Msg("")
		}
	} else {
		l.Info().Msg("migration success")
	}

	return sto
}
