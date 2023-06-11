package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	m "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"pedro.to/rcaptv/database"
)

type Postgres struct {
	db   *sql.DB
	opts *database.StorageOptions
}

func (s *Postgres) Ping(ctx context.Context) (err error) {
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			if err = s.db.Ping(); err == nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Postgres) Migrate() error {
	d, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return err
	}

	mg, err := m.NewWithDatabaseInstance(
		"file://"+s.opts.MigrationPath, "postgres", d,
	)
	if err != nil {
		return err
	}

	return mg.Migrate(uint(s.opts.MigrationVersion))
}

func (s *Postgres) Conn() *sql.DB {
	return s.db
}

func (s *Postgres) Opts() *database.StorageOptions {
	return s.opts
}

func New(opts *database.StorageOptions) database.Storage {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		opts.StorageHost, opts.StoragePort, opts.StorageUser, opts.StoragePassword, opts.StorageDbName,
	))
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(opts.StorageMaxIdleConns)
	db.SetMaxOpenConns(opts.StorageMaxOpenConns)
	db.SetConnMaxLifetime(opts.StorageConnMaxLifetime)

	return &Postgres{
		db:   db,
		opts: opts,
	}
}
