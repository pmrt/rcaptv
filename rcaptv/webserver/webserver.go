package webserver

import (
	"database/sql"

	"golang.org/x/oauth2"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
)

type WebServer struct {
	hx          *helix.Helix
	db          *sql.DB
	tv          *TokenValidator
	oAuthConfig *oauth2.Config
}

type WebServerOpts struct {
	Storage                database.Storage
	ClientID, ClientSecret string
	HelixAPIUrl            string
}

var scopes = []string{"user:read:email"}

// Starts the required services for the WebServer. Make sure to call Shutdown()
func (sv *WebServer) Start() {
	// starts token validator to validate tokens from active users
	go func() {
		sv.tv.Run()
	}()
}

// Stops API services
func (sv *WebServer) Shutdown() {
	sv.tv.Stop()
}

func New(opts WebServerOpts) *WebServer {
	db := opts.Storage.Conn()
	sv := &WebServer{
		oAuthConfig: cfg.OAuthConfig(),
		hx: helix.NewWithUserTokens(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     opts.ClientID,
				ClientSecret: opts.ClientSecret,
			},
			APIUrl: opts.HelixAPIUrl,
		}),
		tv: NewTokenValidator(db, helix.NewWithoutExchange(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     "",
				ClientSecret: "",
			},
			APIUrl: "",
		})),
	}
	return sv
}
