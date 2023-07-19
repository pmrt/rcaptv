package webserver

import (
	"context"
	"database/sql"
	"strconv"
	"time"

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
}

func (v *TokenValidator) AddUser(id int64) {
	v.balancer.Add(strconv.FormatInt(id, 10))
}

func (v *TokenValidator) RemoveUser(id int64) {
	v.balancer.Remove(strconv.FormatInt(id, 10))
}

func (v *TokenValidator) Reset() {
	v.ctx, v.cancel = context.WithCancel(context.Background())
}

func (v *TokenValidator) Run() error {
	v.Reset()
	l := log.With().Str("ctx", "token_validator").Logger()
	l.Info().Msgf("initializing token validator (cycle:1h, estimated_active_users:%d)", v.balancer.EstimatedObjects())
	v.balancer.Start()
	for {
		select {
		case m := <-v.balancer.RealTime():
			for _, usrid := range m.Objects {
				idstr := usrid // keep ref to str for parsing errors
				usrid, err := strconv.ParseInt(idstr, 10, 64)
				if err != nil {
					l.Err(err).Msgf("validator: error while parsing usrid:%s. conv string -> int64 failed (%s)", idstr, err.Error())
					continue
				}

				tks, err := repo.TokenPair(v.db, repo.TokenPairParams{
					UserID: usrid,
				})
				if err != nil {
					l.Err(err).Msgf("validator: error while fetching token pair for usrid:%d (%s)", usrid, err.Error())
				}

				allInvalid := true
				for _, tk := range tks {
					if v.hx.ValidToken(tk.AccessToken) {
						allInvalid = false
						continue
					}

					// invalid token
					if err := repo.DeleteToken(v.db, &repo.DeleteTokenParams{
						UserID:          usrid,
						AccessToken:     tk.AccessToken,
						DeleteUnexpired: true,
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

		case <-v.ctx.Done():
			l.Info().Msg("stopping validator")
			return v.ctx.Err()
		}
	}
}

func (v *TokenValidator) Stop() {
	v.balancer.Cancel()
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
			Freq:             freq,
		}),
		db: db,
		hx: hx,
	}
}
