package repo

import (
	"database/sql"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"golang.org/x/oauth2"
	"pedro.to/rcaptv/gen/tracker/public/model"
	tbl "pedro.to/rcaptv/gen/tracker/public/table"
)

type TokenPairParams struct {
	UserID      int64
	AccessToken string
}

func TokenPair(db *sql.DB, p TokenPairParams) ([]*oauth2.Token, error) {
	if p.UserID == 0 {
		return []*oauth2.Token{}, errors.New("repo.TokenPair: missing user id")
	}

	nowPlus10s := time.Now().Add(10 * time.Second)
	stmt := SELECT(
		tbl.TokenPairs.AllColumns,
	).FROM(tbl.TokenPairs)

	where := tbl.TokenPairs.UserID.EQ(Int(p.UserID)).AND(
		tbl.TokenPairs.ExpiresAt.GT(TimestampT(nowPlus10s)),
	)
	if p.AccessToken != "" {
		where = where.AND(tbl.TokenPairs.AccessToken.EQ(String(p.AccessToken)))
	}
	stmt = stmt.WHERE(where)

	var res []*model.TokenPairs
	err := stmt.Query(db, &res)
	if err != nil {
		return []*oauth2.Token{}, err
	}
	tks := make([]*oauth2.Token, 0, len(res))
	for _, t := range res {
		tks = append(tks, &oauth2.Token{
			AccessToken:  t.AccessToken,
			RefreshToken: t.RefreshToken,
			TokenType:    "Bearer",
			Expiry:       t.ExpiresAt,
		})
	}
	return tks, err
}

func UpsertTokenPair(db *sql.DB, userID int64, t *oauth2.Token) error {
	stmt := tbl.TokenPairs.INSERT(
		tbl.TokenPairs.UserID, tbl.TokenPairs.AccessToken, tbl.TokenPairs.RefreshToken,
		tbl.TokenPairs.ExpiresAt,
	).VALUES(
		userID, t.AccessToken, t.RefreshToken, t.Expiry,
	).ON_CONFLICT(tbl.TokenPairs.UserID, tbl.TokenPairs.RefreshToken).DO_UPDATE(
		SET(
			tbl.TokenPairs.AccessToken.SET(tbl.TokenPairs.EXCLUDED.AccessToken),
			tbl.TokenPairs.ExpiresAt.SET(tbl.TokenPairs.EXCLUDED.ExpiresAt),
			tbl.TokenPairs.LastModifiedAt.SET(TimestampExp(NOW())),
		))
	res, err := stmt.Exec(db)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func DeleteExpired(db *sql.DB) error {
	nowPlus10s := time.Now().Add(10 * time.Second)
	stmt := tbl.TokenPairs.DELETE().
		WHERE(tbl.TokenPairs.ExpiresAt.LT(TimestampT(nowPlus10s)))
	res, err := stmt.Exec(db)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRowsAffected
	}
	return nil
}
