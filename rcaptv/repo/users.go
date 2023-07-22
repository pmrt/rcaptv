package repo

import (
	"database/sql"
	"time"

	. "github.com/go-jet/jet/v2/postgres"

	"pedro.to/rcaptv/gen/tracker/public/model"
	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/helix"
)

func UpsertUser(db *sql.DB, user *helix.User) (int64, error) {
	if user.BroadcasterType == "" {
		user.BroadcasterType = "none"
	}
	stmt := tbl.Users.INSERT(
		tbl.Users.TwitchUserID, tbl.Users.Username,
		tbl.Users.DisplayUsername, tbl.Users.Email, tbl.Users.PpURL,
		tbl.Users.BcType, tbl.Users.TwitchCreatedAt,
	).VALUES(
		user.Id, user.Login, user.DisplayName, user.Email,
		user.ProfileImageURL, user.BroadcasterType, time.Time(user.CreatedAt),
	).ON_CONFLICT(tbl.Users.TwitchUserID).DO_UPDATE(
		SET(
			tbl.Users.Username.SET(tbl.Users.EXCLUDED.Username),
			tbl.Users.DisplayUsername.SET(tbl.Users.EXCLUDED.DisplayUsername),
			tbl.Users.Email.SET(
				StringExp(COALESCE(NULLIF(tbl.Users.EXCLUDED.Email, String("")), tbl.Users.Email)),
			),
			tbl.Users.PpURL.SET(tbl.Users.EXCLUDED.PpURL),
			tbl.Users.BcType.SET(tbl.Users.EXCLUDED.BcType),
			tbl.Users.LastLoginAt.SET(TimestampExp(NOW())),
		)).RETURNING(tbl.Users.UserID)

	var returned []int64
	err := stmt.Query(db, &returned)
	if err != nil {
		return -1, err
	}
	if len(returned) == 0 {
		return -1, ErrNoRowsAffected
	}
	return returned[0], nil
}

type UserQueryParams struct {
	UserID       int64
	TwitchUserID string
	Username     string
}

func User(db *sql.DB, p UserQueryParams) (*model.Users, error) {
	stmt := tbl.Users.SELECT(
		tbl.Users.AllColumns,
	).FROM(tbl.Users)
	if p.UserID != 0 {
		stmt = stmt.WHERE(tbl.Users.UserID.EQ(Int(p.UserID)))
	}
	if p.TwitchUserID != "" {
		stmt = stmt.WHERE(tbl.Users.TwitchUserID.EQ(String(p.TwitchUserID)))
	}
	if p.Username != "" {
		stmt = stmt.WHERE(tbl.Users.Username.EQ(String(p.Username)))
	}
	stmt = stmt.LIMIT(1)

	var u model.Users
	return &u, stmt.Query(db, &u)
}

// ActiveUsers return users that have non-expired tokens
func ActiveUsers(db *sql.DB) ([]*model.Users, error) {
	nowPlus10s := time.Now().Add(10 * time.Second)
	stmt := SELECT(
		tbl.Users.AllColumns,
	).
		DISTINCT(tbl.Users.UserID).
		FROM(
			tbl.Users.INNER_JOIN(
				tbl.TokenPairs,
				tbl.TokenPairs.UserID.EQ(tbl.Users.UserID),
			),
		).
		WHERE(
			tbl.TokenPairs.ExpiresAt.GT(TimestampT(nowPlus10s)),
		)
	var r []*model.Users
	if err := stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}
