package repo

import (
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"

	"pedro.to/rcaptv/gen/tracker/public/model"
	. "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/logger"
)

// Tracked fetches tracked channels
func Tracked(db *sql.DB) (r []*model.TrackedChannels, err error) {
	l := logger.New("", "query:Tracked")

	stmt := SELECT(
		TrackedChannels.AllColumns,
	).FROM(TrackedChannels)

	if err = stmt.Query(db, &r); err != nil {
		l.Error().Err(err).Msg("error while executing query")
		return r, err
	}
	return r, nil
}
