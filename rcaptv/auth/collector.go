package auth

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"pedro.to/rcaptv/repo"
)

type CollectorCtx struct {
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan struct{}
}

type TokenCollector struct {
	db   *sql.DB
	freq time.Duration
	// useful for testing
	lastCollected int64

	ctx *CollectorCtx
}

func (tc *TokenCollector) Collect() (int64, error) {
	return repo.DeleteToken(tc.db, nil)
}

func (tc *TokenCollector) Run() {
	l := log.With().Str("ctx", "token_collector").Logger()
	tc.resetContext(false)
	ticker := time.NewTicker(tc.freq)
	defer ticker.Stop()

	l.Info().Msgf("initializing token collector (cycle:%.0fmin)", tc.freq.Minutes())
	var (
		err error
		n   int64
	)
	for {
		select {
		case <-tc.context().Done():
			l.Info().Msg("token collector stopped")
			tc.resetContext(true)
			close(tc.ctx.stopping)
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

func (tc *TokenCollector) context() context.Context {
	tc.ctx.mu.Lock()
	defer tc.ctx.mu.Unlock()
	return tc.ctx.ctx
}

func (tc *TokenCollector) resetContext(empty bool) {
	tc.ctx.mu.Lock()
	defer tc.ctx.mu.Unlock()
	if empty {
		tc.ctx.ctx, tc.ctx.cancel = nil, nil
	} else {
		tc.ctx.ctx, tc.ctx.cancel = context.WithCancel(context.Background())
		tc.ctx.stopping = make(chan struct{})
	}
}

// Stop the collector. Stop is idempotent
func (tc *TokenCollector) Stop() {
	tc.ctx.mu.Lock()
	if tc.ctx.ctx != nil && tc.ctx.cancel != nil {
		tc.ctx.cancel()
	}
	tc.ctx.mu.Unlock()

	<-tc.ctx.stopping
}

func NewCollector(db *sql.DB, freq time.Duration) *TokenCollector {
	return &TokenCollector{
		db:   db,
		freq: freq,
		ctx:  new(CollectorCtx),
	}
}
