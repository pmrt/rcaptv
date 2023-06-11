package repo

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"pedro.to/rcaptv/database"
	pg "pedro.to/rcaptv/database/postgres"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Run a docker with a database for testing
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.3-alpine3.16",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=user",
			"POSTGRES_DB=name",
			"listen_addresses = '*'",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		panic(err)
	}
	res.Expire(120)

	// Prepare a connection to the db in the docker
	sto := database.New(
		pg.New(&database.StorageOptions{
			StorageHost:            res.GetBoundIP("5432/tcp"),
			StoragePort:            res.GetPort("5432/tcp"),
			StorageUser:            "user",
			StoragePassword:        "test",
			StorageDbName:          "name",
			StorageMaxIdleConns:    5,
			StorageMaxOpenConns:    10,
			StorageConnMaxLifetime: time.Hour,
			StorageConnTimeout:     60 * time.Second,
			DebugMode:              true,

			MigrationVersion: 1,
			MigrationPath:    "../database/postgres/migrations",
		}))
	db = sto.Conn()

	// Run tests
	code := m.Run()

	if err := pool.Purge(res); err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}
