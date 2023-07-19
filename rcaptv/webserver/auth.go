package webserver

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
// webapp (redirect everything to / except /login, etc)

func (sv *WebServer) WithAuth(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	if creds.IsEmpty() {
		// clear just in case it is corrupt or the user cookie is still there
		sv.clearAuthCookies(c)
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
		OAuthConfig: sv.oAuthConfig,
		Notify: func(t *oauth2.Token) error {
			return sv.onTokenRefresh(c, t)
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
func (sv *WebServer) setSessionCookies(c *fiber.Ctx, p SessionCookiesParams) error {
	var id int64
	if p.UserID == 0 {
		id = UserID(c)
		if id == 0 {
			sv.clearAuthCookies(c)
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
		usr, err := repo.User(sv.db, repo.UserQueryParams{
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

func (sv *WebServer) clearAuthCookies(c *fiber.Ctx) {
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
func (sv *WebServer) ValidateSession(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	id := creds.GetInt64(cookie.UserId)
	if id == 0 {
		sv.clearAuthCookies(c)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	at := creds.Get(cookie.AccessToken)
	if repo.ValidToken(sv.db, id, at) {
		return c.SendStatus(fiber.StatusOK)
	}
	// try to get user with invalid access token and current context with
	// tokenSource. If refresh token is still valid this should work, we would
	// get the user and update both: the user and the token
	resp, err := sv.hx.User(&helix.UserParams{
		Context: c.Context(),
	})
	if err != nil {
		if errors.Is(err, helix.ErrUnauthorized) {
			// refresh token failed. Maybe twitch revoked it or user disconnected
			// our app. Clear cookies. The token collector will get rid of the
			// expired tokens in the database
			sv.clearAuthCookies(c)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if len(resp.Data) != 1 {
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}

	if err := sv.upsertUser(c, &resp.Data[0], &oauth2.Token{
		AccessToken:  at,
		RefreshToken: creds.Get(cookie.RefreshToken),
		Expiry:       creds.GetTime(cookie.Expiry),
		TokenType:    "Bearer",
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("couldn't update user")
	}
	return c.SendStatus(fiber.StatusOK)
}

// upsertSession adds or updates the session. An existing user in the database
// is required.
//
// Usr is not required, but provide it if it is already available from where
// this method is called, it saves a roundtrip to the database.
func (sv *WebServer) upsertSession(c *fiber.Ctx, t *oauth2.Token, usrid int64, usr *helix.User) error {
	if usrid == 0 {
		sv.clearAuthCookies(c)
		return errors.New("corrupt user")
	}
	// upsert token pair in database
	if err := repo.UpsertTokenPair(sv.db, usrid, t); err != nil {
		return err
	}
	// update token and session data in cookies
	if err := sv.setSessionCookies(c, SessionCookiesParams{
		Token:  t,
		UserID: usrid,
		User:   usr,
	}); err != nil {
		return err
	}
	// update token validator schedule. We consider an user active when the
	// session is updated
	sv.tv.AddUser(usrid)
	return nil
}

// upsertUser adds or updates the user
func (sv *WebServer) upsertUser(c *fiber.Ctx, usr *helix.User, t *oauth2.Token) error {
	// upsert user in database
	id, err := repo.UpsertUser(sv.db, usr)
	if err != nil {
		return err
	}
	return sv.upsertSession(c, t, id, usr)
}

// onTokenRefresh is called when the token is being refreshed. Generally from
// the tokenSource when the token expired before/in the middle of a request.
// Assumes an existing user in the database. Errors will propagate back to the
// http client
func (sv *WebServer) onTokenRefresh(c *fiber.Ctx, t *oauth2.Token) error {
	return sv.upsertSession(c, t, UserID(c), nil)
}

func (sv *WebServer) Login(c *fiber.Ctx) error {
	if err := sv.ValidateSession(c); err == nil {
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

	return c.Redirect(sv.oAuthConfig.AuthCodeURL(state), fiber.StatusTemporaryRedirect)
}

func (sv *WebServer) Callback(c *fiber.Ctx) error {
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

	tk, err := sv.oAuthConfig.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ctx := helix.ContextWithTokenSource(tk, helix.NotifyReuseTokenSourceOpts{
		OAuthConfig: sv.oAuthConfig,
		// omit notify since we just got a new token pair and we don't expect it to
		// be refreshed during this request.
	})
	resp, err := sv.hx.User(&helix.UserParams{
		Context: ctx,
	})
	if err != nil || len(resp.Data) != 1 {
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}
	if err := sv.upsertUser(c, &resp.Data[0], tk); err != nil {
		sv.clearAuthCookies(c)
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
