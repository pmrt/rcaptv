package auth

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/scheduler"
)

type ReadyCh chan struct{}

type ValidatorCtx struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

type TokenValidator struct {
	balancer *scheduler.BalancedSchedule
	db       *sql.DB
	hx       *helix.Helix
	ctx      *ValidatorCtx

	AfterCycle func(m scheduler.RealTimeMinute)
	readyCh    ReadyCh

	l zerolog.Logger
}

func (v *TokenValidator) AddUser(id int64) {
	v.balancer.Add(strconv.FormatInt(id, 10))
	if !cfg.IsProd {
		l := log.With().Str("ctx", "token_validator").Logger()
		l.Debug().Msgf(
			"validator: added usrid:%d@%smin",
			id, v.balancer.BalancedMin(strconv.FormatInt(id, 10)))
	}
}

func (v *TokenValidator) RemoveUser(id int64) {
	v.balancer.Remove(strconv.FormatInt(id, 10))
	if !cfg.IsProd {
		l := log.With().Str("ctx", "token_validator").Logger()
		l.Debug().Msgf(
			"validator: removed usrid:%d@%smin",
			id, v.balancer.BalancedMin(strconv.FormatInt(id, 10)))
	}
}

func (v *TokenValidator) resetContext(empty bool) {
	v.ctx.mu.Lock()
	defer v.ctx.mu.Unlock()
	if empty {
		v.ctx.ctx, v.ctx.cancel = nil, nil
	} else {
		v.ctx.ctx, v.ctx.cancel = context.WithCancel(context.Background())
	}
}

func (v *TokenValidator) context() context.Context {
	v.ctx.mu.Lock()
	defer v.ctx.mu.Unlock()
	return v.ctx.ctx
}

// Ready returns a channel that will be closed when ready. It can be used to
// block before the validator is ready
func (v *TokenValidator) Ready() <-chan struct{} {
	return v.readyCh
}

func (v *TokenValidator) Run() error {
	v.resetContext(false)
	v.l = log.With().Str("ctx", "token_validator").Logger()
	v.l.Info().Msgf("initializing token validator (cycle:%dmin estimated_active_users:%d)", cycleSize, v.balancer.EstimatedObjects())
	v.l.Info().Msg("validator: retrieving current active users")
	usrs, err := repo.ActiveUsers(v.db)
	if err != nil {
		v.l.Panic().Msg(err.Error())
	}
	for _, usr := range usrs {
		v.AddUser(int64(usr.UserID))
	}
	v.l.Info().Msgf("validator: added:%d (balancer:%p)", len(usrs), v.balancer)

	ready := make(chan struct{}, 1)

	v.balancer.Start()
	for {
		select {
		case m := <-v.balancer.RealTime():
			v.cycle(m)
		case ready <- struct{}{}:
			// notify we're ready. readyCh is a channel that no goroutine will ever
			// write, so it is safe to check if it is closed by trying to receive
			// from it
			select {
			case <-v.readyCh:
				// if closed this would never block and therefore never enter default
			default:
				// if not closed <-v.readyCh will block and enter here
				close(v.readyCh)
			}

		case <-v.context().Done():
			ctx := v.context()
			v.resetContext(true)
			v.l.Info().Msg("stopping validator")
			return ctx.Err()
		}
	}
}

func (v *TokenValidator) cycle(m scheduler.RealTimeMinute) {
	if !cfg.IsProd {
		v.l.Debug().Msgf("validator: min:%s objs:%v", m.Min, m.Objects)
	}
	ctx := v.context()
	for _, usrid := range m.Objects {
		// NOTE: this path may become very slow if we have lots of concurrent
		// active users that need validation. The roundtrip to twitch API is the
		// slowest part here. If scheduler starts to warn about realTime minute
		// discards we should make this faster. Maybe parallel queries? If that's
		// not enough we could factorize validator service into multiple instances
		// and distribute queries across instances
		idstr := usrid // keep ref to str for parsing errors
		usrid, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			v.l.Err(err).Msgf("validator: error while parsing usrid:%s. conv string -> int64 failed (%s)", idstr, err.Error())
			continue
		}

		if !cfg.IsProd {
			v.l.Debug().Msgf("validator: validate usrid:%s", idstr)
		}

		tks, err := repo.TokenPair(v.db, repo.TokenPairParams{
			UserID:  usrid,
			Context: ctx,
		})
		if err != nil {
			v.l.Err(err).Msgf("validator: error while fetching token pair for usrid:%d (%s)", usrid, err.Error())
		}

		allInvalid := true
		for _, tk := range tks {
			if v.hx.ValidToken(helix.ValidTokenParams{
				AccessToken: tk.AccessToken,
				Context:     ctx,
			}) {
				allInvalid = false
				continue
			}

			// invalid token
			if _, err := repo.DeleteToken(v.db, &repo.DeleteTokenParams{
				UserID:          usrid,
				AccessToken:     tk.AccessToken,
				DeleteUnexpired: true,
				Context:         ctx,
			}); err != nil {
				if err == repo.ErrNoRowsAffected {
					v.l.Err(err).Msgf("validator: no rows affected but wanted to delete token for usrid:%d", usrid)
					continue
				}
				v.l.Err(err).Msgf("validator: could not delete token for usrid: %d (%s)", usrid, err.Error())
				continue
			}
		}

		if allInvalid {
			// we only are interested in keep validating active users
			v.RemoveUser(usrid)
		}
	}
	v.AfterCycle(m)
}

func (v *TokenValidator) Stop() {
	v.balancer.Stop()

	v.ctx.mu.Lock()
	defer v.ctx.mu.Unlock()
	if v.ctx.ctx != nil && v.ctx.cancel != nil {
		v.ctx.cancel()
	}
}

var (
	cycleSize = uint(60)
	freq      = time.Minute
)

func NewTokenValidator(db *sql.DB, hx *helix.Helix) *TokenValidator {
	return &TokenValidator{
		balancer: scheduler.New(scheduler.BalancedScheduleOpts{
			CycleSize:        cycleSize,
			EstimatedObjects: uint(cfg.EstimatedActiveUsers),
			BalanceStrategy:  scheduler.StrategyMurmur(uint32(cycleSize)),
			Freq:             freq,
			Salt:             cfg.BalancerSalt,
		}),
		db:         db,
		hx:         hx,
		AfterCycle: func(m scheduler.RealTimeMinute) {},
		readyCh:    make(ReadyCh, 1),
		ctx:        new(ValidatorCtx),
	}
}
