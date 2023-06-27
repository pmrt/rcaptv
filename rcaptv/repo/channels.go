package repo

import (
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"

	"pedro.to/rcaptv/gen/tracker/public/model"
	tbl "pedro.to/rcaptv/gen/tracker/public/table"
)

// Tracked fetches tracked channels
func Tracked(db *sql.DB) (r []*model.TrackedChannels, err error) {
	stmt := SELECT(
		tbl.TrackedChannels.AllColumns,
	).FROM(tbl.TrackedChannels)

	if err = stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}