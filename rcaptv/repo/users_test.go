package repo

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"pedro.to/rcaptv/helix"
)

func TestUpsertUser(t *testing.T) {
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

	cleanupUserAndTokens()
}
