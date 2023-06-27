package tracker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/test"
)

var db *sql.DB

func TestMain(m *testing.M) {
	conn, pool, res := test.SetupPostgres()
	db = conn

	// Run tests
	code := m.Run()

	if err := test.CancelPostgres(pool, res); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)

}

func TestFetchVods(t *testing.T) {
	t.Parallel()
	vodsJson := []byte(`{"data":[{"created_at":"2023-06-14T23:21:38Z","description":"","duration":"1h6m40s","id":"1846472757","language":"es","muted_segments":null,"published_at":"2023-06-14T23:21:38Z","stream_id":"46935025356","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM","type":"archive","url":"https://www.twitch.tv/videos/1846472757","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":7865,"viewable":"public"},{"created_at":"2023-06-14T07:21:23Z","description":"","duration":"5h55m50s","id":"1845909865","language":"es","muted_segments":null,"published_at":"2023-06-14T07:21:23Z","stream_id":"46932511084","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡","type":"archive","url":"https://www.twitch.tv/videos/1845909865","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":39695,"viewable":"public"},{"created_at":"2023-06-13T07:21:20Z","description":"","duration":"8h34m10s","id":"1845060937","language":"es","muted_segments":null,"published_at":"2023-06-13T07:21:20Z","stream_id":"39674758421","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e50667e1d13ad2e09b4f_zeling_39674758421_1686640875//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2  RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡 stream CORTO","type":"archive","url":"https://www.twitch.tv/videos/1845060937","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":53928,"viewable":"public"},{"created_at":"2023-06-12T17:05:21Z","description":"","duration":"2h9m50s","id":"1844500473","language":"es","muted_segments":null,"published_at":"2023-06-12T17:05:21Z","stream_id":"39673161157","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/2da3c11328ff9ec12e1d_zeling_39673161157_1686589516//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844500473","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":17792,"viewable":"public"},{"created_at":"2023-06-12T06:37:33Z","description":"","duration":"7h56m10s","id":"1844222546","language":"es","muted_segments":null,"published_at":"2023-06-12T06:37:33Z","stream_id":"39671417045","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e4dae1bdae5f2a390746_zeling_39671417045_1686551846//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844222546","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":58274,"viewable":"public"}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"}}`)
	wantQuery := "first=1&period=week&type=archive&user_id=58753574"
	bid := "58753574"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(vodsJson)
	}))
	defer sv.Close()

	hx := helix.NewWithoutExchange(&helix.HelixOpts{
		APIUrl: sv.URL,
	}, sv.Client())
	tracker := &Tracker{
		ctx:               context.Background(),
		hx:                hx,
		lastVIDByStreamer: make(lastVODTable, 20),
	}

	// test empty lastVODs table
	_, err := tracker.FetchVods(bid)
	if err != nil {
		t.Fatal(err)
	}
	want := "1846472757"
	if got := tracker.lastVIDByStreamer[bid]; got != want {
		t.Fatalf("expected lastVOD to be %s, got %s", want, got)
	}

	// test again with a older lastVOD in the table
	wantQuery = "first=100&period=week&type=archive&user_id=58753574"
	tracker.lastVIDByStreamer[bid] = "1845060937"
	vods, err := tracker.FetchVods(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(vods) != 3 {
		t.Fatalf("expected exactly 3 vods, got %d", len(vods))
	}
	wantVods := []string{"1846472757", "1845909865", "1845060937"}
	for i, vod := range vods {
		got := vod.VideoID
		want := wantVods[i]
		if got != want {
			t.Fatalf("expected vod %d to be %s, got %s", i, want, got)
		}
	}
}

func TestTrackerStop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	tracker := New(&TrackerOpts{
		Helix:                nil,
		Context:              ctx,
		Storage:              nil,
		TrackingCycleMinutes: 720,
	})
	tracker.FakeRun = true
	tracker.db = db

	timeout := time.After(3 * time.Second)
	actionTimeout := time.After(1 * time.Second)
	done := make(chan bool)
	go func() {
		if err := tracker.Run(); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Fatal(err)
			}
		}
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("tracker did not stop")
	case <-actionTimeout:
		cancel()
	case <-done:
		if !tracker.stopped {
			t.Fatal("tracker did not stop")
		}
	}
}
