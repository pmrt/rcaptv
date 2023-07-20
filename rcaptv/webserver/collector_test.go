package webserver

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

func TestTokenCollectorRun(t *testing.T) {
	defer cleanupUserAndTokens()
	twitchCreatedAt, err := time.Parse(time.RFC3339, "2015-05-02T17:47:43Z")
	if err != nil {
		t.Fatal(err)
	}
	id, err := repo.UpsertUser(db, &helix.User{
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
	// 2 valid tokens
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1_VALID",
		RefreshToken: "REFRESH1",
		Expiry:       time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2_VALID",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	tc := NewCollector(db, time.Millisecond*20)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		tc.Run()
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 30)
	tc.Stop()
	wg.Wait()

	tks, err := repo.TokenPair(db, repo.TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 2 {
		t.Fatalf("expected 2 valid tokens, got %d", len(tks))
	}
	if tc.lastCollected != 0 {
		t.Fatalf("expected token collector to have collected 0 tokens, got:%d", tc.lastCollected)
	}

	// 2 invalid tokens
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS3_VALID",
		RefreshToken: "REFRESH3",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS4_VALID",
		RefreshToken: "REFRESH4",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		tc.Run()
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 30)
	tc.Stop()
	wg.Wait()
	tks, err = repo.TokenPair(db, repo.TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 2 {
		t.Fatalf("expected 2 valid tokens, got %d", len(tks))
	}
	if tc.lastCollected != 2 {
		t.Fatalf("expected token collector to have collected 2 tokens, got:%d", tc.lastCollected)
	}
}
