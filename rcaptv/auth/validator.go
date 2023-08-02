package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/scheduler"
)

type TokenValidator struct {
	balancer *scheduler.BalancedSchedule
	db       *sql.DB
	hx       *helix.Helix
	ctx      context.Context
	cancel   context.CancelFunc

	AfterCycle func(m scheduler.RealTimeMinute)
}

func (v *TokenValidator) AddUser(id int64) {
	v.balancer.Add(strconv.FormatInt(id, 10))
	if !cfg.IsProd {
		l := log.With().Str("ctx", "token_validator").Logger()
		l.Debug().Msgf(
			"validator: added usrid:%d@%smin",
			id, v.balancer.BalancedMin(strconv.FormatInt(id, 10)))
		fmt.Printf("\nAddUser\nschedule:\n%s\nkeyToMin:\n%s\n(balancer:%p)\n",
			spew.Sdump(v.balancer.TestSchedule()), spew.Sdump(v.balancer.TestKeyToMinute()), v.balancer,
		)
	}
}

func (v *TokenValidator) RemoveUser(id int64) {
	v.balancer.Remove(strconv.FormatInt(id, 10))
	if !cfg.IsProd {
		l := log.With().Str("ctx", "token_validator").Logger()
		l.Debug().Msgf(
			"validator: removed usrid:%d@%smin",
			id, v.balancer.BalancedMin(strconv.FormatInt(id, 10)))
		fmt.Printf("\nAddUser\nschedule:\n%s\nkeyToMin:\n%s\n(balancer:%p)\n",
			spew.Sdump(v.balancer.TestSchedule()), spew.Sdump(v.balancer.TestKeyToMinute()), v.balancer,
		)
	}
}

func (v *TokenValidator) Reset() {
	v.ctx, v.cancel = context.WithCancel(context.Background())
}

func (v *TokenValidator) Run() error {
	v.Reset()
	l := log.With().Str("ctx", "token_validator").Logger()
	l.Info().Msgf("initializing token validator (cycle:%dmin estimated_active_users:%d)", cycleSize, v.balancer.EstimatedObjects())
	l.Info().Msg("validator: retrieving current active users")
	usrs, err := repo.ActiveUsers(v.db)
	if err != nil {
		l.Panic().Msg(err.Error())
	}
	for _, usr := range usrs {
		v.AddUser(int64(usr.UserID))
	}
	l.Info().Msgf("validator: added:%d (balancer:%p)", len(usrs), v.balancer)

	v.balancer.Start()
	for {
		select {
		case m := <-v.balancer.RealTime():
			if !cfg.IsProd {
				l.Debug().Msgf("validator: min:%s objs:%v", m.Min, m.Objects)
				fmt.Printf("\nValidator (balancer:%p)\nschedule:\n%s\nkeyToMin:\n%s",
					v.balancer, spew.Sdump(v.balancer.TestSchedule()), spew.Sdump(v.balancer.TestKeyToMinute()),
				)
			}
			for _, usrid := range m.Objects {
				idstr := usrid // keep ref to str for parsing errors
				usrid, err := strconv.ParseInt(idstr, 10, 64)
				if err != nil {
					l.Err(err).Msgf("validator: error while parsing usrid:%s. conv string -> int64 failed (%s)", idstr, err.Error())
					continue
				}

				if !cfg.IsProd {
					l.Debug().Msgf("validator: validate usrid:%s", idstr)
				}

				tks, err := repo.TokenPair(v.db, repo.TokenPairParams{
					UserID:  usrid,
					Context: v.ctx,
				})
				if err != nil {
					l.Err(err).Msgf("validator: error while fetching token pair for usrid:%d (%s)", usrid, err.Error())
				}

				allInvalid := true
				for _, tk := range tks {
					if v.hx.ValidToken(helix.ValidTokenParams{
						AccessToken: tk.AccessToken,
						Context:     v.ctx,
					}) {
						allInvalid = false
						continue
					}

					// invalid token
					if _, err := repo.DeleteToken(v.db, &repo.DeleteTokenParams{
						UserID:          usrid,
						AccessToken:     tk.AccessToken,
						DeleteUnexpired: true,
						Context:         v.ctx,
					}); err != nil {
						if err == repo.ErrNoRowsAffected {
							l.Err(err).Msgf("validator: no rows affected but wanted to delete token for usrid:%d", usrid)
							continue
						}
						l.Err(err).Msgf("validator: could not delete token for usrid: %d (%s)", usrid, err.Error())
						continue
					}
				}

				if allInvalid {
					// we only are interested in keep validating active users
					v.RemoveUser(usrid)
				}
			}
			v.AfterCycle(m)

		case <-v.ctx.Done():
			l.Info().Msg("stopping validator")
			return v.ctx.Err()
		}
	}
}

func (v *TokenValidator) Stop() {
	v.balancer.Stop()
	v.cancel()
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
	}
}
