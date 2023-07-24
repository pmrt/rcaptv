package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"

	"pedro.to/rcaptv/repo"
)

type TokenCollector struct {
	db     *sql.DB
	freq   time.Duration
	ctx    context.Context
	cancel context.CancelFunc
	// useful for testing
	lastCollected int64
}

func (tc *TokenCollector) Collect() (int64, error) {
	return repo.DeleteToken(tc.db, nil)
}

func (tc *TokenCollector) Run() {
	l := log.With().Str("ctx", "token_collector").Logger()
	tc.ctx, tc.cancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(tc.freq)
	defer ticker.Stop()

	l.Info().Msgf("initializing token collector (cycle:%.0fmin)", tc.freq.Minutes())
	var (
		err error
		n   int64
	)
	for {
		select {
		case <-tc.ctx.Done():
			tc.ctx = nil
			tc.cancel = nil
			l.Info().Msg("token collector stopped")
			return
		case <-ticker.C:
			if n, err = tc.Collect(); err != nil {
				if err != repo.ErrNoRowsAffected {
					l.Err(err).Msgf("token collector: could not collect tokens, '%s'", err.Error())
				}
			}
			tc.lastCollected = n
			l.Info().Msgf("token collector: collected:%d", n)
		}
	}
}

// Stop the collector. Stop is idempotent
func (tc *TokenCollector) Stop() {
	if tc.ctx != nil && tc.cancel != nil {
		tc.cancel()
	}
}

func NewCollector(db *sql.DB, freq time.Duration) *TokenCollector {
	return &TokenCollector{
		db:   db,
		freq: freq,
	}
}
