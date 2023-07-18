package repo

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/oauth2"

	"pedro.to/rcaptv/helix"
)

func TestUpsertAndSelectTokenPair(t *testing.T) {
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

	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS123",
		RefreshToken: "REFRESH456",
		Expiry:       time.Now().Add(time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	tks, err := TokenPair(db, TokenPairParams{
		UserID: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 1 {
		t.Fatalf("expected 1 token, got: %d", len(tks))
	}
	got, want := tks[0].AccessToken, "ACCESS123"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS789",
		RefreshToken: "REFRESH456",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	tks, err = TokenPair(db, TokenPairParams{
		UserID: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 1 {
		t.Fatalf("expected 1 token (the same token should be updated), got:%d", len(tks))
	}
	got, want = tks[0].AccessToken, "ACCESS789"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	got, want = tks[0].RefreshToken, "REFRESH456"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	cleanupUserAndTokens()
}

func TestTokenPairAccessToken(t *testing.T) {
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

	// 2 valid tokens for first user
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	// 2 valid tokens + 1 invalid for second user (test target)
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS3",
		RefreshToken: "REFRESH3",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS4",
		RefreshToken: "REFRESH4",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS5",
		RefreshToken: "REFRESH5",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	tks, err := TokenPair(db, TokenPairParams{
		UserID:      id2,
		AccessToken: "ACCESS5",
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want := len(tks), 0
	if got != want {
		t.Fatalf("expected %d invalid tokens, got %d", want, got)
	}

	wantToken := "ACCESS4"
	tks, err = TokenPair(db, TokenPairParams{
		UserID:      id2,
		AccessToken: wantToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, want = len(tks), 1
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d", want, got)
	}
	gotToken := tks[0].AccessToken
	if gotToken != wantToken {
		t.Fatalf("got %q, want %q", gotToken, wantToken)
	}
	cleanupUserAndTokens()
}

func TestDeleteExpired(t *testing.T) {
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

	// 2 valid tokens + 1 invalid for first user
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(time.Hour * 2),
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
	// 2 invalid tokens + 1 valid for second user
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS4",
		RefreshToken: "REFRESH4",
		Expiry:       time.Now().Add(time.Hour * 2),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS5",
		RefreshToken: "REFRESH5",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id2, &oauth2.Token{
		AccessToken:  "ACCESS6",
		RefreshToken: "REFRESH6",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	if err := DeleteToken(db, nil); err != nil {
		t.Fatal(err)
	}

	tks, err := TokenPair(db, TokenPairParams{
		UserID: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want := len(tks), 2
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d", want, got)
	}

	tks, err = TokenPair(db, TokenPairParams{
		UserID: id2,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want = len(tks), 1
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d", want, got)
	}
	cleanupUserAndTokens()
}

func TestDeleteExpiredSingle(t *testing.T) {
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
	// 4 invalid tokens
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now(),
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
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS4",
		RefreshToken: "REFRESH4",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	if err := DeleteToken(db, &DeleteTokenParams{
		UserID:      id,
		AccessToken: "ACCESS4",
	}); err != nil {
		t.Fatal(err)
	}

	tks, err := TokenPair(db, TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want := len(tks), 3
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d", want, got)
	}

	if err := DeleteToken(db, &DeleteTokenParams{
		UserID:       id,
		RefreshToken: "REFRESH1",
	}); err != nil {
		t.Fatal(err)
	}

	tks, err = TokenPair(db, TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want = len(tks), 2
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d", want, got)
	}

	gotToken, wantToken := tks[0].AccessToken, "ACCESS2"
	if gotToken != wantToken {
		t.Fatalf("expected token %s, got %s", wantToken, gotToken)
	}
	gotToken, wantToken = tks[1].AccessToken, "ACCESS3"
	if gotToken != wantToken {
		t.Fatalf("expected token %s, got %s", wantToken, gotToken)
	}
	cleanupUserAndTokens()
}

func TestDeleteValid(t *testing.T) {
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
	// 1 valid token + 1 invalid token
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now().Add(4 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	if err := DeleteToken(db, &DeleteTokenParams{
		UserID:      id,
		AccessToken: "ACCESS1",
	}); err != nil && err != ErrNoRowsAffected {
		t.Fatal(err)
	}

	tks, err := TokenPair(db, TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want := len(tks), 2
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d. Valid token should not be deleted with DeleteUnexpired=false", want, got)
	}

	if err := DeleteToken(db, &DeleteTokenParams{
		UserID:          id,
		AccessToken:     "ACCESS1",
		DeleteUnexpired: true,
	}); err != nil {
		t.Fatal(err)
	}

	tks, err = TokenPair(db, TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(tks)
	got, want = len(tks), 1
	if got != want {
		t.Fatalf("expected %d valid tokens, got %d. Valid token should be deleted with DeleteUnexpired=true", want, got)
	}

	gotToken, wantToken := tks[0].AccessToken, "ACCESS2"
	if gotToken != wantToken {
		t.Fatalf("expected token %s, got %s", wantToken, gotToken)
	}
	cleanupUserAndTokens()
}
