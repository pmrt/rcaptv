package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"

	"pedro.to/rcaptv/auth/pb"
	"pedro.to/rcaptv/certs"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/cookie"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

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

type PassportOps struct {
	Storage                database.Storage
	ClientID, ClientSecret string
	HelixAPIUrl            string
}

type rpcClient struct {
	c    pb.TokenValidatorServiceClient
	conn *grpc.ClientConn
}

// Passport handles authentication, session validation and provides some session APIs.
//
// Passport is thread safe.
type Passport struct {
	db          *sql.DB
	oAuthConfig *oauth2.Config
	hx          *helix.Helix

	rpc *rpcClient
}

func (p *Passport) Start() {
}

func (p *Passport) Stop() {
	p.rpc.conn.Close()
}

func (p *Passport) ValidatorAddUser(id int64) {
	_, err := p.rpc.c.AddUser(context.Background(), &pb.User{Id: id})
	if err != nil {
		l := log.With().Str("ctx", "auth").Logger()
		l.Err(err).Msgf("RPCTokenValidator returned an error:'%s' (userid:%d)", err.Error(), id)
	}
}

// note: internal APIs that don't use twitch api will need a different
// middleware that uses validateSession
func (p *Passport) WithAuth(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	if !creds.ValidShape() {
		if !cfg.IsProd {
			l := log.With().Str("ctx", "auth").Logger()
			msg := fmt.Sprintf("invalid token (usrid:%d", creds.GetInt64(cookie.UserId))
			if creds.Empty() {
				msg += " empty"
			}
			if creds.EmptyUser() {
				msg += " emptyUser"
			}
			msg += ")"
			l.Debug().Msg(msg)
		}
		ctx := context.WithValue(c.Context(), CtxKeyLoggedIn, false)
		c.SetUserContext(ctx)
		return c.Next()
	}

	if !cfg.IsProd {
		if creds.Expired() {
			l := log.With().Str("ctx", "auth").Logger()
			l.Debug().Msgf("invalid token (usrid:%d expired)", creds.GetInt64(cookie.UserId))
		}
	}

	t := &oauth2.Token{
		AccessToken:  creds.Get(cookie.AccessToken),
		RefreshToken: creds.Get(cookie.RefreshToken),
		Expiry:       creds.GetTime(cookie.Expiry),
		TokenType:    "Bearer",
	}
	// add tokenSource to Ctx to use and refresh tokens in the request stack
	ctx := helix.ContextWithTokenSource(c.Context(), t, helix.NotifyReuseTokenSourceOpts{
		OAuthConfig: p.oAuthConfig,
		// this functions holds the reference to Fiber.Ctx. The tokensource
		// lifespan is meant to last at least the same as the request lifespan,
		// since it is used within the current request's http.Clients calls right
		// before Do() calls
		Notify: func(t *oauth2.Token) error {
			if !cfg.IsProd {
				l := log.With().Str("ctx", "auth").Logger()
				l.Debug().Msgf("refreshing token for usrid:%d", UserID(c))
			}
			return p.onTokenRefresh(c, t)
		},
	})
	// make user id and token pair available too in the request stack
	ctx = context.WithValue(ctx, CtxKeyUserID, creds.GetInt64(cookie.UserId))
	ctx = context.WithValue(ctx, CtxKeyToken, t)
	ctx = context.WithValue(ctx, CtxKeyLoggedIn, true)
	c.SetUserContext(ctx)
	return c.Next()
}

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
func (p *Passport) setSessionCookies(c *fiber.Ctx, params SessionCookiesParams) error {
	var id int64
	if params.UserID == 0 {
		id = UserID(c)
		if id == 0 {
			ClearAuthCookies(c)
			return errors.New("corrupt user")
		}
	}

	creds := cookie.New()
	creds.
		Add(cookie.AccessToken, params.Token.AccessToken).
		Add(cookie.RefreshToken, params.Token.RefreshToken).
		AddTime(cookie.Expiry, params.Token.Expiry).
		AddInt64(cookie.UserId, params.UserID)

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
	if params.User == nil {
		usr, err := repo.User(p.db, repo.UserQueryParams{
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
		usrCookie.Add(cookie.TwitchID, params.User.Id).
			Add(cookie.Username, params.User.Login).
			Add(cookie.DisplayName, params.User.DisplayName).
			Add(cookie.ProfilePicture, params.User.ProfileImageURL).
			Add(cookie.BcType, params.User.BroadcasterType)
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

// ValidateSession is and endpoint to be invoked from time to time, generally
// asynchronously at the startup of the client.
//
// Makes a request to /users endpoint to update tokens and user info in both
// database and cookies.
//
// It assumes that the UserID and tokenSource is already in the request context,
// so call it after the WithAuth middleware.
func (p *Passport) ValidateSession(c *fiber.Ctx) error {
	creds := cookie.Fiber(c, cookie.CredentialsCookie)
	if !creds.ValidShape() {
		ClearAuthCookies(c)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	id := creds.GetInt64(cookie.UserId)
	at := creds.Get(cookie.AccessToken)
	if !creds.Expired() && repo.ValidToken(p.db, c.Context(), id, at) {
		// skip further validation if the token is still in the db
		return c.SendStatus(fiber.StatusOK)
	}
	// try to get user with invalid access token and current context with
	// tokenSource. If refresh token is still valid this should work, we would
	// get the user and update both: the user and the token
	resp, err := p.hx.User(&helix.UserParams{
		Context: c.UserContext(),
	})
	if err != nil {
		if errors.Is(err, helix.ErrUnauthorized) {
			// refresh token failed. Maybe twitch revoked it or user disconnected
			// our app. Clear cookies. The token collector will get rid of the
			// expired tokens in the database
			ClearAuthCookies(c)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if len(resp.Data) != 1 {
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}

	if err := p.upsertUser(c, &resp.Data[0], &oauth2.Token{
		AccessToken:  at,
		RefreshToken: creds.Get(cookie.RefreshToken),
		Expiry:       creds.GetTime(cookie.Expiry),
		TokenType:    "Bearer",
	}); err != nil {
		ClearAuthCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("couldn't update user")
	}
	return c.SendStatus(fiber.StatusOK)
}

// upsertSession adds or updates the session. An existing user in the database
// is required.
//
// Usr is not required, but provide it if it is already available from where
// this method is called, it saves a roundtrip to the database.
func (p *Passport) upsertSession(c *fiber.Ctx, t *oauth2.Token, usrid int64, usr *helix.User) error {
	if usrid == 0 {
		ClearAuthCookies(c)
		return errors.New("corrupt user")
	}
	// upsert token pair in database
	if err := repo.UpsertTokenPair(p.db, usrid, t); err != nil {
		return err
	}
	// update token and session data in cookies
	if err := p.setSessionCookies(c, SessionCookiesParams{
		Token:  t,
		UserID: usrid,
		User:   usr,
	}); err != nil {
		return err
	}
	// update token validator schedule. We consider an user active when the
	// session is updated
	p.ValidatorAddUser(usrid)
	return nil
}

// upsertUser adds or updates the user
func (p *Passport) upsertUser(c *fiber.Ctx, usr *helix.User, t *oauth2.Token) error {
	// upsert user in database
	id, err := repo.UpsertUser(p.db, usr)
	if err != nil {
		return err
	}
	return p.upsertSession(c, t, id, usr)
}

// onTokenRefresh is called when the token is being refreshed. Generally from
// the tokenSource when the token expired before/in the middle of a request.
// Assumes an existing user in the database. Errors will propagate back to the
// http client
func (p *Passport) onTokenRefresh(c *fiber.Ctx, t *oauth2.Token) error {
	return p.upsertSession(c, t, UserID(c), nil)
}

var secret = genSecret

func (p *Passport) Login(c *fiber.Ctx) error {
	err := p.ValidateSession(c)
	status := c.Response().StatusCode()
	if err == nil && status == fiber.StatusOK {
		// session is valid, do nothing
		return c.Redirect("/", fiber.StatusTemporaryRedirect)
	}

	state, err := secret(80)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookie.OauthStateCookie,
		Value:    state,
		Expires:  timeNow().Add(30 * time.Minute),
		Path:     cfg.AuthEndpoint,
		Domain:   cfg.Domain,
		SameSite: "Lax",
		HTTPOnly: true,
		Secure:   true,
	})
	return c.Redirect(p.oAuthConfig.AuthCodeURL(state), fiber.StatusTemporaryRedirect)
}

func (p *Passport) Callback(c *fiber.Ctx) error {
	state := c.Cookies(cookie.OauthStateCookie)
	if state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("missing state challenge")
	}
	if state != c.Query("state") {
		return c.Status(fiber.StatusBadRequest).SendString("invalid state challenge")
	}
	queryErr := c.Query("error")
	if queryErr == "access_denied" {
		ClearAuthorizationCodeCookies(c)
		return c.Render("access_denied", nil)
	} else if queryErr == "redirect_mismatch" {
		ClearAuthorizationCodeCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("invalid redirect")
	} else if queryErr != "" {
		ClearAuthorizationCodeCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("invalid authorization")
	}

	tk, err := p.oAuthConfig.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		ClearAuthorizationCodeCookies(c)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ctx := helix.ContextWithTokenSource(c.Context(), tk, helix.NotifyReuseTokenSourceOpts{
		OAuthConfig: p.oAuthConfig,
		// omit notify since we just got a new token pair and we don't expect it to
		// be refreshed during this request.
	})
	resp, err := p.hx.User(&helix.UserParams{
		Context: ctx,
	})
	if err != nil || len(resp.Data) != 1 {
		ClearAuthorizationCodeCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("invalid user")
	}
	if err := p.upsertUser(c, &resp.Data[0], tk); err != nil {
		ClearAllCookies(c)
		return c.Status(fiber.StatusInternalServerError).SendString("failed to create user")
	}
	return c.Redirect("/", fiber.StatusTemporaryRedirect)
}

func (p *Passport) Logout(c *fiber.Ctx) error {
	id := UserID(c)
	if id == 0 {
		return c.Redirect("/", fiber.StatusTemporaryRedirect)
	}
	t, ok := c.UserContext().Value(CtxKeyToken).(*oauth2.Token)
	if !ok {
		return c.Redirect("/", fiber.StatusTemporaryRedirect)
	}

	params := &repo.DeleteTokenParams{
		UserID:          id,
		DeleteUnexpired: true,
		RefreshToken:    t.RefreshToken,
		Context:         c.Context(),
	}
	// ?all=1 -> Remove all tokens for the user
	if c.Query("all") == "1" {
		params.RefreshToken = ""
	}
	if _, err := repo.DeleteToken(p.db, params); err != nil {
		return c.Redirect("/", fiber.StatusTemporaryRedirect)
	}
	ClearAllCookies(c)
	return c.Redirect("/", fiber.StatusTemporaryRedirect)
}

// NewRPCClient returns a rpcClient. gRPC.ClientConn closing must be handled.
func NewRPCClient() (*rpcClient, error) {
	creds, err := credentials.NewClientTLSFromFile(certs.Filename("ca_cert.pem"), "")
	if err != nil {
		panic("could not create gRPC credentials: " + err.Error())
	}
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", cfg.RPCAuthHost, cfg.RPCAuthPort),
		grpc.WithTransportCredentials(creds),
		// grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	c := pb.NewTokenValidatorServiceClient(conn)
	return &rpcClient{
		conn: conn,
		c:    c,
	}, nil
}

func ping(rpc *rpcClient, ctx context.Context) (err error) {
	l := log.With().Str("ctx", "auth").Logger()
	timer := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-timer.C:
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if _, err = rpc.c.Ping(timeoutCtx, &emptypb.Empty{}); err == nil {
				cancel()
				return
			}
			cancel()
			l.Info().Msg("retrying ping...")
		case <-ctx.Done():
			return errors.New("could not establish connection with RPC server " + err.Error())
		}
	}
}

// creates a new Passport. Stop must be handled.
func New(opts PassportOps) *Passport {
	l := log.With().Str("ctx", "auth").Logger()

	l.Info().Msgf("passport: creating RPC client")
	rpc, err := NewRPCClient()
	if err != nil {
		panic("grpc error: " + err.Error())
	}
	l.Info().Msgf("ping RPC server @ %s:%s", cfg.RPCAuthHost, cfg.RPCAuthPort)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := ping(rpc, ctx); err != nil {
		panic("ping RPCTokenValidator: " + err.Error())
	}
	l.Info().Msg("connection to RPC server successful")
	return &Passport{
		db:          opts.Storage.Conn(),
		oAuthConfig: cfg.OAuthConfig(),
		hx: helix.NewWithUserTokens(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     opts.ClientID,
				ClientSecret: opts.ClientSecret,
			},
			APIUrl: opts.HelixAPIUrl,
		}),
		rpc: rpc,
	}
}
