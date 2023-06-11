package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	LogLevel int8
)

func SetupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.SetGlobalLevel(zerolog.Level(LogLevel))
}

func New(cmd, ctx string) zerolog.Logger {
	logger := log.With().
		Str("cmd", cmd).
		Str("ctx", ctx).
		Logger()

	return logger
}
