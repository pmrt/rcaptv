package repo

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/oauth2"

	"pedro.to/rcaptv/helix"
)

func TestUpsertUser(t *testing.T) {
	defer cleanupUserAndTokens()
	twitchCreatedAt, err := time.Parse(time.RFC3339, "2015-05-02T17:47:43Z")
	if err != nil {
		t.Fatal(err)
	}
	shouldBeIgnoredTs, err := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	id, err := UpsertUser(db, &helix.User{
		Id:              "90075649",
		Login:           "illojuan",
		DisplayName:     "IlloJuan",
		Email:           "test@email.com",
		ProfileImageURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png",
		BroadcasterType: "partner",
		CreatedAt:       helix.RFC3339Timestamp(twitchCreatedAt),
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := User(db, UserQueryParams{
		TwitchUserID: "90075649",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(u))
	if u.Email != "test@email.com" {
		t.Fatalf("expected email to be test@email.com, got: %s", u.Email)
	}
	createdAt := u.CreatedAt

	id2, err := UpsertUser(db, &helix.User{
		Id:              "90075649",
		Login:           "illojuan2",
		DisplayName:     "IlloJuan",
		Email:           "test@email.com2",
		ProfileImageURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png",
		BroadcasterType: "partner",
		CreatedAt:       helix.RFC3339Timestamp(shouldBeIgnoredTs),
	})
	if err != nil {
		t.Fatal(err)
	}

	// should be updated not inserted, so expect same id
	got, want := id2, id
	if got != want {
		t.Fatalf("got %d want %d", got, want)
	}

	u, err = User(db, UserQueryParams{
		TwitchUserID: "90075649",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(u))
	if u.Username != "illojuan2" {
		t.Fatalf("expected username to be updated, expect: illojuan2, got: %s", u.Username)
	}
	if u.Email != "test@email.com2" {
		t.Fatalf("expected email to be updated, expect: test@email.com2, got: %s", u.Email)
	}
	if *u.CreatedAt != *createdAt {
		t.Fatalf("created_at should not be updated, expect: %s, got: %s", createdAt, u.CreatedAt)
	}
}

func TestActiveUsers(t *testing.T) {
	defer cleanupUserAndTokens()
	twitchCreatedAt, err := time.Parse(time.RFC3339, "2015-05-02T17:47:43Z")
	if err != nil {
		t.Fatal(err)
	}
	id, err := UpsertUser(db, &helix.User{
		Id:              "90075649",
		Login:           "illojuan",
		DisplayName:     "IlloJuan",
		Email:           "test@email.com",
		ProfileImageURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png",
		BroadcasterType: "partner",
		CreatedAt:       helix.RFC3339Timestamp(twitchCreatedAt),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 2 valids + 1 invalid
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now().Add(time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS3",
		RefreshToken: "REFRESH3",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	twitchCreatedAt2, err := time.Parse(time.RFC3339, "2015-05-02T20:47:43Z")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := UpsertUser(db, &helix.User{
		Id:              "90075650",
		Login:           "illojuan2",
		DisplayName:     "IlloJuan2",
		Email:           "test@email.com",
		ProfileImageURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png",
		BroadcasterType: "partner",
		CreatedAt:       helix.RFC3339Timestamp(twitchCreatedAt2),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 1 Invalid
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS4",
		RefreshToken: "REFRESH4",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	id3, err := UpsertUser(db, &helix.User{
		Id:              "90075651",
		Login:           "illojuan3",
		DisplayName:     "IlloJuan3",
		Email:           "test@email.com",
		ProfileImageURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png",
		BroadcasterType: "partner",
		CreatedAt:       helix.RFC3339Timestamp(twitchCreatedAt2),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 1 Valid
	if err = UpsertTokenPair(db, id3, &oauth2.Token{
		AccessToken:  "ACCESS5",
		RefreshToken: "REFRESH5",
		Expiry:       time.Now().Add(2 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	usrs, err := ActiveUsers(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(usrs) != 2 {
		t.Fatalf("expected 2 user, got:%d", len(usrs))
	}
	got := usrs[0].TwitchUserID
	want := "90075649"
	if got != want {
		t.Fatalf("expected twitch_user_id:%s, got:%s", want, got)
	}
	got = usrs[1].TwitchUserID
	want = "90075651"
	if got != want {
		t.Fatalf("expected twitch_user_id:%s, got:%s", want, got)
	}
}
