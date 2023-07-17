package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"pedro.to/rcaptv/helix"
)

func TestSetSessionCookies(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	api := &API{
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
		if err := api.setSessionCookies(c, params); err != nil {
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
