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

	// whether to include invalid tokens (default:false)
	Invalid bool
}

// TokenPair returns the corresponding oauth2.Token objects in the database for
// the required UserID parameter. If an optional AccessToken parameter is
// given, it will try to find the given access token for the given userID
// returning an empty slice otherwise.
//
// By default TokenPair filters out non-expired tokens. If Invalid is true, it
// also will include expired tokens.
func TokenPair(db *sql.DB, p TokenPairParams) ([]*oauth2.Token, error) {
	if p.UserID == 0 {
		return []*oauth2.Token{}, errors.New("repo.TokenPair: missing user id")
	}

	nowPlus10s := time.Now().Add(10 * time.Second)
	stmt := SELECT(
		tbl.TokenPairs.AllColumns,
	).FROM(tbl.TokenPairs)
	where := tbl.TokenPairs.UserID.EQ(Int(p.UserID))
	if p.AccessToken != "" {
		where = where.AND(tbl.TokenPairs.AccessToken.EQ(String(p.AccessToken)))
	}
	if !p.Invalid {
		where = where.AND(tbl.TokenPairs.ExpiresAt.GT(TimestampT(nowPlus10s)))
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

// ValidTokens returns whether the given access token for the provided userId
// is found on the database and it is not expired
func ValidToken(db *sql.DB, userID int64, accessToken string) bool {
	tks, err := TokenPair(db, TokenPairParams{
		UserID:      userID,
		AccessToken: accessToken,
	})
	if err != nil {
		return false
	}
	if len(tks) == 0 {
		return false
	}
	return tks[0].AccessToken == accessToken
}

// UpserTokenPair takes an userID and a given oauth2.Token, if the refresh
// tokens exists for the userID it will update the access token and the expiry,
// otherwise it will insert a new token pair
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

type DeleteTokenParams struct {
	UserID       int64
	AccessToken  string
	RefreshToken string

	// If DeleteUnexpired is true. Valid tokens will be deleted too (default:false)
	DeleteUnexpired bool
}

// DeleteToken deletes expired tokens matching the provided parameters. If a
// nil parameter object is passed, all expired tokens will be deleted.
//
// IF DeleteUnexpired=true is passed down in the params, DeleteToken will delete
// matching non-expired tokens too
func DeleteToken(db *sql.DB, p *DeleteTokenParams) error {
	nowPlus10s := time.Now().Add(10 * time.Second)
	stmt := tbl.TokenPairs.DELETE()
	where := tbl.TokenPairs.ExpiresAt.LT(TimestampT(nowPlus10s))
	if p != nil {
		if p.DeleteUnexpired {
			where = Bool(true)
		}
		if p.UserID != 0 {
			where = where.AND(tbl.TokenPairs.UserID.EQ(Int(p.UserID)))
		}
		if p.AccessToken != "" {
			where = where.AND(tbl.TokenPairs.AccessToken.EQ(String(p.AccessToken)))
		}
		if p.RefreshToken != "" {
			where = where.AND(tbl.TokenPairs.RefreshToken.EQ(String(p.RefreshToken)))
		}
	}
	stmt = stmt.WHERE(where)
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
