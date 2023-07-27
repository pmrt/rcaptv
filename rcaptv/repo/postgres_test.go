package repo

import (
	"database/sql"
	"log"
	"os"
	"testing"

	jet "github.com/go-jet/jet/v2/postgres"

	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/test"
)

var db *sql.DB

func cleanupUserAndTokens() {
	stmt := tbl.TokenPairs.DELETE().WHERE(jet.Bool(true))
	if _, err := stmt.Exec(db); err != nil {
		log.Fatal(err)
	}
	stmt = tbl.Users.DELETE().WHERE(jet.Bool(true))
	if _, err := stmt.Exec(db); err != nil {
		log.Fatal(err)
	}
}

func cleanupClips() {
	stmt := tbl.Clips.DELETE().WHERE(jet.Bool(true))
	if _, err := stmt.Exec(db); err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	conn, pool, res := test.SetupPostgres()
	db = conn

	// Run tests
	code := m.Run()

	if err := test.CancelPostgres(pool, res); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
