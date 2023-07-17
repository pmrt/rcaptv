package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

type CtxKeyAPI int

const CtxKeyUserID CtxKeyAPI = iota

func UserID(c *fiber.Ctx) int64 {
	if v, ok := c.Context().Value(CtxKeyUserID).(int64); ok {
		return v
	}
	return 0
}

// Token collector service (3d?)
// Token validator service (1h)
// webapp (redirect everything to / except /login, etc)

// servicio para borrar tokens exp cada día
// servicio cada hora para validar tokens
// borrar cookies
// añadir cookies
// meter en ctx el tokenSource con WithAuth

func (a *API) WithAuth(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	if creds.IsEmpty() {
		// clear just in case it is corrupt or the user cookie is still there
		a.clearAuthCookies(c)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	t := &oauth2.Token{
		AccessToken:  creds.Get(cookie.AccessToken),
		RefreshToken: creds.Get(cookie.RefreshToken),
		Expiry:       creds.GetTime(cookie.Expiry),
		TokenType:    "Bearer",
	}
	// add tokenSource to Ctx to use and refresh tokens in the request stack
	ctx := helix.ContextWithTokenSource(t, helix.NotifyReuseTokenSourceOpts{
		OAuthConfig: a.OAuthConfig,
		Notify: func(t *oauth2.Token) error {
			return a.onTokenRefresh(c, t)
		},
	})
	// make user id available too in the request stack
	ctx = context.WithValue(ctx, CtxKeyUserID, creds.GetInt64(cookie.UserId))
	c.SetUserContext(ctx)
	return c.Next()
}

type SessionCookiesParams struct {
	Token  *oauth2.Token
	UserID int64
	User   *helix.User
}

type UserCookie struct {
	TwitchID          string `json:"twitch_id"`
	Username          string `json:"username"`
	DisplayName       string `json:"display_name"`
	ProfilePictureURL string `json:"profile_picture_url"`
	BcType            string `json:"bc_type"`
}

var timeNow = time.Now

// setSessionCookies sets the cookies needed for user session. Usually updated with
// the database to keep state in sync.
//
// Two cookies are used: CredentialCookie and UserCookie. The former is
// encrypted and non-available within javascript and contains sensible data
// like tokens, id, expiry date, etc. The UserCookie is not encrypted and it's
// available within javascript, it is intended only for visuals (e.g.
// Displaying username when logged in or profile picture)
//
// The only required parameter is the Token and the ID if the api context does
// not contain it. Prefer passing down everything if possible. User will save
// a roundtrip to the database if provided.
func (a *API) setSessionCookies(c *fiber.Ctx, p SessionCookiesParams) error {
	var id int64
	if p.UserID == 0 {
		id = UserID(c)
		if id == 0 {
			a.clearAuthCookies(c)
			return errors.New("corrupt user")
		}
	}

	creds := cookie.New()
	creds.
		Add(cookie.AccessToken, p.Token.AccessToken).
		Add(cookie.RefreshToken, p.Token.RefreshToken).
		AddTime(cookie.Expiry, p.Token.Expiry).
		AddInt64(cookie.UserId, p.UserID)

	exp := timeNow().Add(time.Hour * 24 * 365 * 10)
	c.Cookie(&fiber.Cookie{
		Name:     cookie.CredentialsCookie,
		Value:    creds.String(),
		Path:     "/",
		Expires:  exp,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: true,
		Secure:   true,
	})

	usrCookie := cookie.New()
	if p.User == nil {
		usr, err := repo.User(a.db, repo.UserQueryParams{
			UserID: id,
		})
		if err != nil {
			return errors.New("couldn't retrieve user")
		}
		usrCookie.Add(cookie.TwitchID, usr.TwitchUserID).
			Add(cookie.Username, usr.Username).
			Add(cookie.DisplayName, usr.DisplayUsername).
			Add(cookie.ProfilePicture, *usr.PpURL).
			Add(cookie.BcType, string(usr.BcType))
	} else {
		usrCookie.Add(cookie.TwitchID, p.User.Id).
			Add(cookie.Username, p.User.Login).
			Add(cookie.DisplayName, p.User.DisplayName).
			Add(cookie.ProfilePicture, p.User.ProfileImageURL).
			Add(cookie.BcType, p.User.BroadcasterType)
	}
	// Important: this cookie is not secure. It is httpOnly=false and unencrypted
	// it is accesible from javascript and intended for visual feedback only
	c.Cookie(&fiber.Cookie{
		Name:     cookie.UserCookie,
		Value:    usrCookie.String(),
		Path:     "/",
		Expires:  exp,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: false,
		Secure:   true,
	})
	return nil
}

func (a *API) clearAuthCookies(c *fiber.Ctx) {
	c.ClearCookie(cookie.CredentialsCookie)
	c.ClearCookie(cookie.UserCookie)
}

// ValidateSession is and endpoint to be invoked from time to time, generally
// asynchronously at the startup of the client.
//
// Makes a request to /users endpoint to update tokens and user info in both
// database and cookies.
//
// It assumes that the UserID and tokenSource is already in the request context,
// so call it after the WithAuth middleware.
func (a *API) ValidateSession(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	id := creds.GetInt64(cookie.UserId)
	if id == 0 {
		a.clearAuthCookies(c)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	at := creds.Get(cookie.AccessToken)
	if repo.ValidToken(a.db, id, at) {
		return c.SendStatus(fiber.StatusOK)
	}
	// try to get user with invalid access token and current context with
	// tokenSource. If refresh token is still valid this should work, we would
	// get the user and update both: the user and the token
	resp, err := a.hx.User(&helix.UserParams{
		Context: c.Context(),
	})
	if err != nil {
		if errors.Is(err, helix.ErrUnauthorized) {
			// refresh token failed. Maybe twitch revoked it or user disconnected
			// our app. Clear cookies. The token collector will get rid of the
			// expired tokens in the database
			a.clearAuthCookies(c)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if len(resp.Data) != 1 {
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}

	if err := a.AddOrUpdateUser(c, &resp.Data[0], &oauth2.Token{
		AccessToken:  at,
		RefreshToken: creds.Get(cookie.RefreshToken),
		Expiry:       creds.GetTime(cookie.Expiry),
		TokenType:    "Bearer",
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("couldn't update user")
	}
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) AddOrUpdateUser(c *fiber.Ctx, usr *helix.User, t *oauth2.Token) error {
	// upsert user in database
	id, err := repo.UpsertUser(a.db, usr)
	if err != nil {
		return err
	}
	// upsert token pair in database
	if err := repo.UpsertTokenPair(a.db, id, t); err != nil {
		return err
	}
	// update cookies
	if err := a.setSessionCookies(c, SessionCookiesParams{
		Token:  t,
		UserID: id,
		User:   usr,
	}); err != nil {
		return err
	}
	return nil
}

// onTokenRefresh is called when the token is being refreshed. Generally from
// the tokenSource when the token expired before/in the middle of a request.
// Assumes an existing user in the database. Errors will propagate back to the
// http client
func (a *API) onTokenRefresh(c *fiber.Ctx, t *oauth2.Token) error {
	id := UserID(c)
	if id == 0 {
		a.clearAuthCookies(c)
		return errors.New("corrupt user")
	}
	// update db
	if err := repo.UpsertTokenPair(a.db, id, t); err != nil {
		return err
	}
	// update cookies
	if err := a.setSessionCookies(c, SessionCookiesParams{
		Token:  t,
		UserID: id,
	}); err != nil {
		return err
	}
	return nil
}

func (a *API) Login(c *fiber.Ctx) error {
	if err := a.ValidateSession(c); err == nil {
		// session is valid, do nothing
		return c.Redirect("/", fiber.StatusTemporaryRedirect)
	}

	state, err := genSecret(32)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookie.OauthStateCookie,
		Value:    state,
		Expires:  time.Now().Add(30 * time.Minute),
		Path:     cfg.AuthEndpoint,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: true,
		Secure:   true,
	})

	return c.Redirect(a.OAuthConfig.AuthCodeURL(state), fiber.StatusTemporaryRedirect)
}

func (a *API) Callback(c *fiber.Ctx) error {
	defer func() {
		// whatever happens, clear the state cookie
		c.ClearCookie(cookie.OauthStateCookie)
	}()

	state := c.Cookies(cookie.OauthStateCookie)
	if state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("missing state challenge")
	}
	if state != c.Query("state") {
		return c.Status(fiber.StatusBadRequest).SendString("invalid state challenge")
	}
	if c.Query("error") == "access_denied" {
		return c.Render("views/access_denied", nil)
	}

	tk, err := a.OAuthConfig.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ctx := helix.ContextWithTokenSource(tk, helix.NotifyReuseTokenSourceOpts{
		OAuthConfig: a.OAuthConfig,
		// omit notify since we just got a new token pair and we don't expect it to
		// be refreshed during this request.
	})
	resp, err := a.hx.User(&helix.UserParams{
		Context: ctx,
	})
	if err != nil || len(resp.Data) != 1 {
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}
	if err := a.AddOrUpdateUser(c, &resp.Data[0], tk); err != nil {
		a.clearAuthCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("failed to create user")
	}
	return c.Redirect("/", fiber.StatusTemporaryRedirect)
}

func genSecret(n int) (string, error) {
	k := make([]byte, n)
	if _, err := rand.Read(k); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(k), nil
}
