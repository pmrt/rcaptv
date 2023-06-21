package repo

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"pedro.to/rcaptv/test"
)

var db *sql.DB

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
