package logger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Fiber() fiber.Handler {
	l := log.With().Str("ctx", "api").Logger()
	return func(c *fiber.Ctx) error {
		req := c.Request()
		l.Info().Msgf(
			"[%s] %s: %s %s",
			c.IP(), c.Method(), c.Path(), req.URI().QueryString(),
		)
		return c.Next()
	}
}