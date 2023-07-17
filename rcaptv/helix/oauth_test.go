package helix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNotifyTokenSource(t *testing.T) {
	t.Parallel()
	recv := make([]*oauth2.Token, 0, 2)
	f := func(t *oauth2.Token) error {
		recv = append(recv, t)
		return nil
	}

	authReqs := 1
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/test" {
			w.Write([]byte("ok"))
			return
		}

		if r.URL.String() == "/token" {
			w.Header().Set("content-type", "application/json")
			if authReqs == 1 {
				w.Write([]byte(`{"access_token": "NEW_ACCESS_TOKEN", "scope": "user", "token_type": "Bearer", "expires_in": 1, "refresh_token": "NEW_REFRESH_TOKEN"}`))
			}
			if authReqs == 2 {
				w.Write([]byte(`{"access_token": "NEW_ACCESS_TOKEN2", "scope": "user", "token_type": "Bearer", "expires_in": 86400, "refresh_token": "NEW_REFRESH_TOKEN"}`))
			}
			authReqs++
		}
	}))
	defer sv.Close()

	config := &oauth2.Config{
		ClientID:     "CLIENT_ID123",
		ClientSecret: "CLIENT_SECRET123",
		RedirectURL:  "REDIRECT_URL123",
		Scopes:       []string{"SCOPE1"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  sv.URL + "/auth",
			TokenURL: sv.URL + "/token",
		},
	}
	tk := &oauth2.Token{
		AccessToken:  "OLD_ACCESS_TOKEN",
		TokenType:    "Bearer",
		RefreshToken: "OLD_REFRESH_TOKEN",
		Expiry:       time.Now(),
	}

	c := oauth2.NewClient(context.Background(), NotifyReuseTokenSource(tk, NotifyReuseTokenSourceOpts{
		OAuthConfig: config,
		Notify:      f,
	}))
	if _, err := c.Get(sv.URL + "/test"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	if _, err := c.Get(sv.URL + "/test"); err != nil {
		t.Fatal(err)
	}

	if len(recv) != 2 {
		t.Fatal("expected 2 tokens")
	}
	if recv[0].AccessToken != "NEW_ACCESS_TOKEN" {
		t.Fatal("expected first token to be NEW_ACCESS_TOKEN")
	}
	if recv[1].AccessToken != "NEW_ACCESS_TOKEN2" {
		t.Fatal("expected second token to be NEW_ACCESS_TOKEN2")
	}
}
