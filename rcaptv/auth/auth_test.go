package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"

	"pedro.to/rcaptv/helix"
)

func TestSetSessionCookies(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	p := &Passport{
		db: nil,
		hx: nil,
	}
	ts, err := time.Parse(time.RFC3339, "2015-05-02T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	timeNow = func() time.Time { return ts }
	params := SessionCookiesParams{
		Token: &oauth2.Token{
			AccessToken:  "ACCESS",
			RefreshToken: "REFRESH",
			Expiry:       ts,
			TokenType:    "Bearer",
		},
		UserID: 1,
		User: &helix.User{
			Id:              "100000",
			Login:           "test",
			DisplayName:     "test",
			Type:            "",
			BroadcasterType: "",
			ProfileImageURL: "https://pictureexample.com",
			OfflineImageURL: "https://offlinepicture.com",
			CreatedAt:       helix.RFC3339Timestamp(ts),
			Description:     "description",
			ViewCount:       2,
			Email:           "email@example.com",
		},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		if err := p.setSessionCookies(c, params); err != nil {
			t.Fatal(err)
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) != 2 {
		t.Fatal("Set-Cookie header should have 2 values")
	}

	want := []string{
		"credentials=a=ACCESS&e=2015-05-02T10%3A00%3A00Z&r=REFRESH&u=1; expires=Tue, 29 Apr 2025 10:00:00 GMT; path=/; HttpOnly; secure; SameSite=Lax",
		"user=bc_type=&display_name=test&login=test&profile_picture=https%3A%2F%2Fpictureexample.com&twitch_id=100000; expires=Tue, 29 Apr 2025 10:00:00 GMT; path=/; secure; SameSite=Lax",
	}
	for i, cookie := range cookies {
		if want[i] != cookie {
			t.Fatalf("g=got w=want\ng: '%s'\nw: '%s'", cookie, want[i])
		}
	}
}

func TestValidateSessionInvalidShape(t *testing.T) {
	t.Parallel()

	p := &Passport{}
	app := fiber.New()
	app.Get("/validate", p.ValidateSession)

	req, err := http.NewRequest("GET", "/validate", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", "credentials=a=ACCESS&e=2015-05-02T10%3A00%3A00Z&r=REFRESH&u=0; expires=Thu, 14 Jul 2033 18:13:59 GMT; path=/; HttpOnly; secure; SameSite=Lax; user=bc_type=&display_name=test&login=test&profile_picture=https%3A%2F%2Fpictureexample.com&twitch_id=100000; expires=Thu, 14 Jul 2033 18:13:59 GMT; path=/; secure; SameSite=Lax")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	resetCookies := resp.Header.Values("Set-Cookie")
	if len(resetCookies) != 2 {
		t.Fatalf("expected 2 Set-Cookie headers to reset cookie, got:%d", len(resetCookies))
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got:%d", resp.StatusCode)
	}
}

func TestLoginEmptyCookie(t *testing.T) {
	t.Parallel()
	redirectURL := "http://fakeredirect.com"
	secretText := "fakesecret"
	scope := "read:user:email"
	secret = func(n int) (string, error) {
		return secretText, nil
	}
	timeNow = func() time.Time {
		ts, err := time.Parse(time.RFC3339, "2023-01-01T17:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		return ts
	}

	p := &Passport{
		db: nil,
		oAuthConfig: &oauth2.Config{
			ClientID:     "",
			ClientSecret: "",
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  redirectURL,
			Scopes:       []string{scope},
		},
	}
	app := fiber.New()
	app.Get("/login", p.Login)

	req, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", resp.StatusCode)
	}
	got := resp.Header.Get("Set-Cookie")
	want := fmt.Sprintf(
		"oauth_state=%s; expires=%s; path=/; HttpOnly; secure; SameSite=Lax",
		secretText, timeNow().Add(30*time.Minute).UTC().Format(http.TimeFormat),
	)
	if got != want {
		t.Fatalf("unexpected cookie, want:'%s' got:'%s'", want, got)
	}
	got = resp.Header.Get("Location")
	want = fmt.Sprintf(
		"?client_id=&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(redirectURL), url.QueryEscape(scope), secretText,
	)
	if got != want {
		t.Fatalf("unexpected redirect location want:'%s', got:'%s'", want, got)
	}
}
