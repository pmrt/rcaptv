package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/oauth2"

	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/scheduler"
)

func TestTokenValidatorRequests(t *testing.T) {
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
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1_VALID",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2_EXPIRED",
		RefreshToken: "REFRESH3",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	var wgRequest sync.WaitGroup
	reqs := make([]*http.Request, 0, 2)
	nreqs := 1
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nreqs == 1 {
			t.Log("#1 req")
			reqs = append(reqs, r.Clone(context.Background()))
			w.WriteHeader(200)
			w.Write([]byte(`{"client_id":"wbmytr93xzw8zbg0p1izqyzzc5mbiz","login":"twitchdev","scopes":["channel:read:subscriptions"],"user_id":"141981764","expires_in":5520838}`))
			nreqs++
			wgRequest.Done()
			return
		} else if nreqs == 2 {
			t.Log("#2 req")
			reqs = append(reqs, r.Clone(context.Background()))
			w.WriteHeader(401)
			w.Write([]byte(`{"status":401,"message":"invalid access token"}`))
			nreqs++
			wgRequest.Done()
			return
		} else {
			t.Fatal("unexpected request")
		}
	}))

	cycleSize = 1
	freq = time.Millisecond * 25
	tv := NewTokenValidator(db, helix.NewWithoutExchange(&helix.HelixOpts{
		Creds: helix.ClientCreds{
			ClientID:     "",
			ClientSecret: "",
		},
		APIUrl:           "",
		ValidateEndpoint: sv.URL,
	}))
	tv.skipLoadingUsers = true
	wgRequest.Add(2)
	tv.AddUser(id)
	go func() {
		tv.Run()
	}()
	wgRequest.Wait()
	tv.Stop()

	if len(reqs) != 2 {
		t.Fatalf("expected to receive 2 token validation requests, got %d", len(reqs))
	}
	got := reqs[0].Header.Get("Authorization")
	want := "OAuth ACCESS1_VALID"
	if got != want {
		t.Fatalf("Authorization header: want:'%s' got:'%s'", want, got)
	}
	got = reqs[1].Header.Get("Authorization")
	want = "OAuth ACCESS1_VALID"
	if got != want {
		t.Fatalf("Authorization header: want:'%s' got:'%s'", want, got)
	}
}

func TestTokenValidator(t *testing.T) {
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
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS1_VALID",
		RefreshToken: "REFRESH2",
		Expiry:       time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpsertTokenPair(db, id, &oauth2.Token{
		AccessToken:  "ACCESS2_EXPIRED",
		RefreshToken: "REFRESH3",
		Expiry:       time.Now(),
		TokenType:    "Bearer",
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	nreqs := 1
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nreqs == 1 {
			t.Log("#1 req")
			w.WriteHeader(200)
			w.Write([]byte(`{"client_id":"wbmytr93xzw8zbg0p1izqyzzc5mbiz","login":"twitchdev","scopes":["channel:read:subscriptions"],"user_id":"141981764","expires_in":5520838}`))
			nreqs++
			return
		} else if nreqs == 2 {
			t.Log("#2 req")
			w.WriteHeader(401)
			w.Write([]byte(`{"status":401,"message":"invalid access token"}`))
			nreqs++
			return
		}
	}))

	cycleSize = 1
	freq = time.Millisecond * 25
	afterOp := func(op *scheduler.Op) {
		t.Logf("\nscheduler op: %d\n", op.Typ)
		wg.Done()
	}
	tv := &TokenValidator{
		skipLoadingUsers: true,
		balancer: scheduler.New(scheduler.BalancedScheduleOpts{
			CycleSize:        cycleSize,
			EstimatedObjects: 200,
			BalanceStrategy:  scheduler.StrategyMurmur(uint32(cycleSize)),
			Freq:             freq,
			Salt:             "fake_salt",
			AfterOp:          afterOp,
		}),
		db: db,
		hx: helix.NewWithoutExchange(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     "",
				ClientSecret: "",
			},
			APIUrl:           "",
			ValidateEndpoint: sv.URL,
		}),
		AfterCycle: func(m scheduler.RealTimeMinute) {},
		readyCh:    make(ReadyCh, 1),
		ctx:        new(ValidatorCtx),
	}
	wg.Add(1)
	tv.AddUser(id)
	// wg.Done() are within the hook AfterOp
	go func() {
		tv.Run()
	}()
	wg.Wait() // wait until the user is been aded to schedule
	tv.Stop()

	schedule := tv.balancer.UnsafeSchedule()
	keyToMin := tv.balancer.UnsafeKeyToMinute()
	t.Logf("schedule\n%s", spew.Sdump(schedule))
	t.Logf("keyToMin\n%s", spew.Sdump(keyToMin))

	gotN := len(schedule[scheduler.Minute(0)])
	wantN := 1
	if gotN != wantN {
		t.Fatalf("expected schedule to have %d users, got %d", wantN, gotN)
	}
	gotN = len(keyToMin)
	wantN = 1
	if gotN != wantN {
		t.Fatalf("expected keyToMin to have %d users, got %d", wantN, gotN)
	}

	tks, err := repo.TokenPair(db, repo.TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tks))
	}

	wg.Add(1)
	go func() {
		tv.Run()
	}()
	wg.Wait() // wait until user is removed from schedule
	tv.Stop()

	schedule = tv.balancer.UnsafeSchedule()
	keyToMin = tv.balancer.UnsafeKeyToMinute()
	t.Logf("schedule\n%s", spew.Sdump(schedule))
	t.Logf("keyToMin\n%s", spew.Sdump(keyToMin))

	gotN = len(schedule[scheduler.Minute(0)])
	wantN = 0
	if gotN != wantN {
		t.Fatalf("expected schedule to have %d users, got %d", wantN, gotN)
	}
	gotN = len(keyToMin)
	wantN = 0
	if gotN != wantN {
		t.Fatalf("expected keyToMin to have %d users, got %d", wantN, gotN)
	}

	tks, err = repo.TokenPair(db, repo.TokenPairParams{
		UserID:  id,
		Invalid: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tks) != 1 {
		t.Fatalf("expected 1 tokens, got %d", len(tks))
	}
	got, want := tks[0].AccessToken, "ACCESS2_EXPIRED"
	if got != want {
		t.Fatalf("expected token in db to be '%s', got:'%s' ", want, got)
	}
}
