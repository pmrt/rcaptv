package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/gofiber/fiber/v2"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
)

var ExpiredCookieExpiry = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

type CtxKeyAPI int

const (
	CtxKeyUserID CtxKeyAPI = iota
	CtxKeyToken
	CtxKeyLoggedIn
)

func UserID(c *fiber.Ctx) int64 {
	if v, ok := c.UserContext().Value(CtxKeyUserID).(int64); ok {
		return v
	}
	return 0
}

func IsLoggedIn(c *fiber.Ctx) bool {
	if v, ok := c.UserContext().Value(CtxKeyLoggedIn).(bool); ok {
		return v
	}
	return false
}

func genSecret(n int) (string, error) {
	k := make([]byte, n)
	if _, err := rand.Read(k); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(k), nil
}

func ClearAuthCookies(c *fiber.Ctx) {
	// some browsers requires same attributes that when created
	c.Cookie(&fiber.Cookie{
		Name:     cookie.CredentialsCookie,
		Value:    "deleted",
		Path:     "/",
		Expires:  ExpiredCookieExpiry,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: true,
		Secure:   true,
	})
	c.Cookie(&fiber.Cookie{
		Name:     cookie.UserCookie,
		Value:    "{}",
		Path:     "/",
		Expires:  ExpiredCookieExpiry,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: false,
		Secure:   true,
	})
}

func ClearAuthorizationCodeCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     cookie.OauthStateCookie,
		Value:    "deleted",
		Expires:  ExpiredCookieExpiry,
		Path:     cfg.AuthEndpoint,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: true,
		Secure:   true,
	})
}

func ClearAllCookies(c *fiber.Ctx) {
	ClearAuthCookies(c)
	ClearAuthorizationCodeCookies(c)
}
