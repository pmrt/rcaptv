package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func SetupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

func SetLevel(lvl int8) {
	zerolog.SetGlobalLevel(zerolog.Level(lvl))
}

func New(cmd, ctx string) zerolog.Logger {
	logger := log.With()

	if cmd != "" {
		logger = logger.Str("cmd", cmd)
	}
	if ctx != "" {
		logger = logger.Str("ctx", ctx)
	}

	return logger.Logger()
}
