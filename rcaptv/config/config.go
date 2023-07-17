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

const Version = "0.1.7"
const LastMigrationVersion = 1

var loaded = false

var (
	TrackerPostgresUser     string
	TrackerPostgresPassword string
	RcaptvPostgresUser      string
	RcaptvPostgresPassword  string

	PostgresHost                   string
	PostgresPort                   string
	PostgresDBName                 string
	PostgresMaxIdleConns           int
	PostgresMaxOpenConns           int
	PostgresConnTimeoutSeconds     int
	PostgresConnMaxLifetimeMinutes int
	PostgresMigVersion             int
	PostgresMigPath                string

	HelixClientID     string
	HelixClientSecret string
	TestClientID      string
	TestClientSecret  string
	WebhookSecret     string

	SkipMigrations bool

	Domain                  string
	BaseURL                 string
	AuthEndpoint            string
	AuthRedirectEndpoint    string
	CookieSecret            string
	TwitchAPIUrl            string
	APIPort                 string
	EventSubEndpoint        string
	RateLimitMaxConns       int
	RateLimitExpSeconds     int
	ClipsMaxPeriodDiffHours int

	TrackingCycleMinutes     int
	ClipTrackingWindowHours  int
	ClipTrackingMaxDeepLevel int
	ClipViewThreshold        int
	ClipViewWindowSize       int

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
		if !IsProd {
			// security measure to prevent logging secrets in prod
			l.Debug().
				Str("ctx", "config").
				Msgf("=> [%s]: %v", key, val)
		}
		return val
	}
	return def
}

func LoadVars() {
	l := l.With().
		Str("ctx", "config").
		Logger()

	if err := godotenv.Load(); err != nil {
		l.Warn().
			Err(err).
			Msgf("couldn't load .env file: %s. Using passed environment variables or default values", err.Error())
	}

	l.Info().Msg("reading environment variables")

	PostgresHost = Env("POSTGRES_HOST", "127.0.0.1")
	PostgresPort = Env("POSTGRES_PORT", "5432")
	TrackerPostgresUser = Env("TRACKER_POSTGRES_USER", "tracker")
	TrackerPostgresPassword = Env("TRACKER_POSTGRES_PASSWORD", "unsafepassword")
	RcaptvPostgresUser = Env("RCAPTV_POSTGRES_USER", "rcaptv")
	RcaptvPostgresPassword = Env("RCAPTV_POSTGRES_PASSWORD", "unsafepassword")
	PostgresDBName = Env("POSTGRES_DB_NAME", "tracker")
	PostgresMaxIdleConns = Env("POSTGRES_MAX_IDLE_CONNS", 5)
	PostgresMaxOpenConns = Env("POSTGRES_MAX_OPEN_CONNS", 10)
	PostgresConnMaxLifetimeMinutes = Env("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 60)
	PostgresConnTimeoutSeconds = Env("POSTGRES_CONN_TIMEOUT_SECONDS", 60)
	PostgresMigVersion = Env("POSTGRES_MIG_VERSION", LastMigrationVersion)
	PostgresMigPath = Env("POSTGRES_MIG_PATH", "database/postgres/migrations")

	HelixClientID = Env("HELIX_CLIENT_ID", "fake_client_id")
	HelixClientSecret = Env("HELIX_CLIENT_SECRET", "fake_secret")
	TestClientID = Env("TEST_CLIENT_ID", "fake_client_id")
	TestClientSecret = Env("TEST_CLIENT_SECRET", "fake_secret")
	WebhookSecret = Env("WEBHOOK_SECRET", "fake_secret")

	SkipMigrations = Env("SKIP_MIGRATIONS", false)

	Domain = Env("DOMAIN", "localhost")
	BaseURL = Env("BASE_URL", "http://localhost")
	APIPort = Env("API_PORT", "8080")
	AuthEndpoint = Env("AUTH_ENDPOINT", "/auth")
	AuthRedirectEndpoint = Env("AUTH_REDIRECT_ENDPOINT", "/auth/redirect")
	CookieSecret = Env("COOKIE_SECRET", "unsafe_secret")
	TwitchAPIUrl = Env("TWITCH_API_URL", "https://api.twitch.tv/helix")
	EventSubEndpoint = Env("EVENTSUB_ENDPOINT", "/eventsub")
	RateLimitMaxConns = Env("RATE_LIMIT_MAX_CONNS", 20)
	RateLimitExpSeconds = Env("RATE_LIMIT_EXP_SECONDS", 60)
	ClipsMaxPeriodDiffHours = Env("CLIPS_MAX_PERIOD_DIFF_HOURS", 168)

	TrackingCycleMinutes = Env("TRACKING_CYCLE_MINUTES", 720)
	ClipTrackingWindowHours = Env("CLIP_TRACKING_WINDOW_HOURS", 7*24)
	ClipTrackingMaxDeepLevel = Env("CLIP_TRACKING_MAX_DEEP_LEVEL", 2)
	ClipViewThreshold = Env("CLIP_VIEW_THRESHOLD", 10)
	ClipViewWindowSize = Env("CLIP_VIEW_WINDOW_SIZE", 4)

	Debug = Env("DEBUG", false)
	logger.SetLevel(Env("LOG_LEVEL", int8(zerolog.InfoLevel)))
	if Debug {
		logger.SetLevel(Env("LOG_LEVEL", int8(zerolog.DebugLevel)))
	}
}

func Setup() {
	if loaded {
		return
	}
	logger.SetupLogger()
	LoadVars()
	loaded = true
}
