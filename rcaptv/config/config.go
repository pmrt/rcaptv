package config

import (
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	l "github.com/rs/zerolog/log"
	"pedro.to/rcaptv/logger"
)

const LastMigrationVersion = 1

var (
	PostgresHost                   string
	PostgresPort                   string
	PostgresUser                   string
	PostgresPassword               string
	PostgresDBName                 string
	PostgresMaxIdleConns           int
	PostgresMaxOpenConns           int
	PostgresConnTimeoutSeconds     int
	PostgresConnMaxLifetimeMinutes int
	PostgresMigVersion             int
	PostgresMigPath                string

	HelixClientID    string
	HelixSecret      string
	TestClientID     string
	TestClientSecret string

	SkipMigrations bool

	APIUrl  string
	APIPort string

	TrackIntervalMinutes int

	Debug bool
)

type SupportStringconv interface {
	~int | ~int8 | ~int64 | ~float32 | ~string | ~bool
}

func conv(v string, to reflect.Kind) any {
	var err error

	if to == reflect.String {
		return v
	}

	if to == reflect.Bool {
		if bool, err := strconv.ParseBool(v); err == nil {
			return bool
		}
	}

	if to == reflect.Int {
		if int, err := strconv.Atoi(v); err == nil {
			return int
		}
	}

	if to == reflect.Int8 {
		if i64, err := strconv.ParseInt(v, 10, 8); err == nil {
			return int8(i64)
		}
	}

	if to == reflect.Int64 {
		if i64, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i64
		}
	}

	if to == reflect.Float32 {
		if f32, err := strconv.ParseFloat(v, 32); err == nil {
			return f32
		}
	}

	l.Panic().
		Err(err).
		Str("ctx", "config").
		Msg("")
	return nil
}

func Env[T SupportStringconv](key string, def T) T {
	if v, ok := os.LookupEnv(key); ok {
		val := conv(v, reflect.TypeOf(def).Kind()).(T)
		l.Debug().
			Str("ctx", "config").
			Msgf("=> [%s]: %v", key, val)
		return val
	}
	return def
}

func LoadVars() {
	l := l.With().
		Str("ctx", "config").
		Logger()

	if err := godotenv.Load(); err != nil {
		l.Panic().
			Err(err).
			Msg("couldn't load .env file")
	}

	l.Info().Msg("reading environment variables")

	PostgresHost = Env("POSTGRES_HOST", "127.0.0.1")
	PostgresPort = Env("POSTGRES_PORT", "5432")
	PostgresUser = Env("POSTGRES_USER", "tracker")
	PostgresPassword = Env("POSTGRES_PASSWORD", "unsafepassword")
	PostgresDBName = Env("POSTGRES_DB_NAME", "tracker")
	PostgresMaxIdleConns = Env("POSTGRES_MAX_IDLE_CONNS", 5)
	PostgresMaxOpenConns = Env("POSTGRES_MAX_OPEN_CONNS", 10)
	PostgresConnMaxLifetimeMinutes = Env("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 60)
	PostgresConnTimeoutSeconds = Env("POSTGRES_CONN_TIMEOUT_SECONDS", 60)
	PostgresMigVersion = Env("POSTGRES_MIG_VERSION", LastMigrationVersion)
	PostgresMigPath = Env("POSTGRES_MIG_PATH", "database/postgres/migrations")

	HelixClientID = Env("HELIX_CLIENT_ID", "fake_client_id")
	HelixSecret = Env("HELIX_SECRET", "fake_secret")
	TestClientID = Env("TEST_CLIENT_ID", "fake_client_id")
	TestClientSecret = Env("TEST_CLIENT_SECRET", "fake_secret")

	SkipMigrations = Env("SKIP_MIGRATIONS", false)

	APIPort = Env("API_PORT", "8080")
	APIUrl = Env("API_URL", "https://api.twitch.tv/helix")

	Debug = Env("DEBUG", false)
	logger.SetLevel(Env("LOG_LEVEL", int8(zerolog.InfoLevel)))
	if !IsProd {
		Debug = Env("DEBUG", true)
		logger.SetLevel(Env("LOG_LEVEL", int8(zerolog.DebugLevel)))
	}
}

func Setup() {
	logger.SetupLogger()
	LoadVars()
}
