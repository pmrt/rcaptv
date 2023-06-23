package helix

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestVODDurationSeconds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		duration string
		want     int32
	}{
		{
			duration: "1h20m5s",
			want:     60*60 + 20*60 + 5,
		},
		{
			duration: "55m50s",
			want:     55*60 + 50,
		},
		{
			duration: "1h",
			want:     60 * 60,
		},
	}

	for _, test := range tests {
		vod := &VOD{
			DurationString: test.duration,
		}
		got, err := vod.DurationSeconds()
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Fatalf("DurationSeconds() = %d, want %d", got, test.want)
		}
		if got != vod.Duration {
			t.Fatalf("DurationSeconds() = %d, want %d", got, vod.Duration)
		}
	}
}

func TestHelixVOD(t *testing.T) {
	t.Parallel()
	vodsJson := []byte(`{"data":[{"created_at":"2023-06-14T23:21:38Z","description":"","duration":"1h6m40s","id":"1846472757","language":"es","muted_segments":null,"published_at":"2023-06-14T23:21:38Z","stream_id":"46935025356","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM","type":"archive","url":"https://www.twitch.tv/videos/1846472757","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":7865,"viewable":"public"},{"created_at":"2023-06-14T07:21:23Z","description":"","duration":"5h55m50s","id":"1845909865","language":"es","muted_segments":null,"published_at":"2023-06-14T07:21:23Z","stream_id":"46932511084","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡","type":"archive","url":"https://www.twitch.tv/videos/1845909865","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":39695,"viewable":"public"},{"created_at":"2023-06-13T07:21:20Z","description":"","duration":"8h34m10s","id":"1845060937","language":"es","muted_segments":null,"published_at":"2023-06-13T07:21:20Z","stream_id":"39674758421","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e50667e1d13ad2e09b4f_zeling_39674758421_1686640875//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2  RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡 stream CORTO","type":"archive","url":"https://www.twitch.tv/videos/1845060937","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":53928,"viewable":"public"},{"created_at":"2023-06-12T17:05:21Z","description":"","duration":"2h9m50s","id":"1844500473","language":"es","muted_segments":null,"published_at":"2023-06-12T17:05:21Z","stream_id":"39673161157","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/2da3c11328ff9ec12e1d_zeling_39673161157_1686589516//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844500473","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":17792,"viewable":"public"},{"created_at":"2023-06-12T06:37:33Z","description":"","duration":"7h56m10s","id":"1844222546","language":"es","muted_segments":null,"published_at":"2023-06-12T06:37:33Z","stream_id":"39671417045","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e4dae1bdae5f2a390746_zeling_39671417045_1686551846//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844222546","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":58274,"viewable":"public"}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"}}`)
	wantQuery := "first=100&period=week&type=archive&user_id=58753574"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(vodsJson)
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		c: sv.Client(),
	}
	vods, err := hx.Vods(&VODParams{
		BroadcasterID: "58753574",
		StopAtVODID:   "1844500473",
		Period:        Week,
	})
	if err != nil {
		t.Fatal(err)
	}

	ts1, err := time.Parse(time.RFC3339, "2023-06-14T23:21:38Z")
	if err != nil {
		t.Fatal(err)
	}
	ts2, err := time.Parse(time.RFC3339, "2023-06-14T07:21:23Z")
	if err != nil {
		t.Fatal(err)
	}
	ts3, err := time.Parse(time.RFC3339, "2023-06-13T07:21:20Z")
	if err != nil {
		t.Fatal(err)
	}
	ts4, err := time.Parse(time.RFC3339, "2023-06-12T17:05:21Z")
	if err != nil {
		t.Fatal(err)
	}

	want := []*VOD{
		{
			VideoID:        "1846472757",
			BroadcasterID:  "58753574",
			StreamID:       "46935025356",
			CreatedAt:      ts1,
			PublishedAt:    ts1,
			DurationString: "1h6m40s",
			Lang:           "es",
			Title:          "☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      7865,
			Duration:       4000,
		},
		{
			VideoID:        "1845909865",
			BroadcasterID:  "58753574",
			StreamID:       "46932511084",
			CreatedAt:      ts2,
			PublishedAt:    ts2,
			DurationString: "5h55m50s",
			Lang:           "es",
			Title:          "☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      39695,
			Duration:       21350,
		},
		{
			VideoID:        "1845060937",
			BroadcasterID:  "58753574",
			StreamID:       "39674758421",
			CreatedAt:      ts3,
			PublishedAt:    ts3,
			DurationString: "8h34m10s",
			Lang:           "es",
			Title:          "☣️DROPS☣️BELLUM I TIER 2  RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡 stream CORTO",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e50667e1d13ad2e09b4f_zeling_39674758421_1686640875//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      53928,
			Duration:       30850,
		},
		{
			VideoID:        "1844500473",
			BroadcasterID:  "58753574",
			StreamID:       "39673161157",
			CreatedAt:      ts4,
			PublishedAt:    ts4,
			DurationString: "2h9m50s",
			Lang:           "es",
			Title:          "☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/2da3c11328ff9ec12e1d_zeling_39673161157_1686589516//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      17792,
			Duration:       7790,
		},
	}

	for i, vod := range vods {
		if diff := deep.Equal(want[i], vod); diff != nil {
			t.Fatal(diff)
		}
	}
}

func TestHelixVODEmpty(t *testing.T) {
	t.Parallel()
	vodsJson := []byte(`{"data":[],"pagination":{}}`)
	wantQuery := "first=100&period=week&type=archive&user_id=58753574"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(vodsJson)
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		c: sv.Client(),
	}
	vods, err := hx.Vods(&VODParams{
		BroadcasterID: "58753574",
		StopAtVODID:   "1844500473",
		Period:        Week,
	})
	if err != nil {
		if !errors.Is(err, ErrItemsEmpty) {
			t.Fatal(err)
		}
	} else {
		t.Fatal("expected ErrItemsEmpty error")
	}

	if len(vods) != 0 {
		t.Fatal("expected empty vods")
	}
}

func TestHelixVODOnlyMostRecent(t *testing.T) {
	t.Parallel()
	vodsJson := []byte(`{"data":[{"created_at":"2023-06-14T23:21:38Z","description":"","duration":"1h6m40s","id":"1846472757","language":"es","muted_segments":null,"published_at":"2023-06-14T23:21:38Z","stream_id":"46935025356","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM","type":"archive","url":"https://www.twitch.tv/videos/1846472757","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":7865,"viewable":"public"},{"created_at":"2023-06-14T07:21:23Z","description":"","duration":"5h55m50s","id":"1845909865","language":"es","muted_segments":null,"published_at":"2023-06-14T07:21:23Z","stream_id":"46932511084","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡","type":"archive","url":"https://www.twitch.tv/videos/1845909865","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":39695,"viewable":"public"},{"created_at":"2023-06-13T07:21:20Z","description":"","duration":"8h34m10s","id":"1845060937","language":"es","muted_segments":null,"published_at":"2023-06-13T07:21:20Z","stream_id":"39674758421","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e50667e1d13ad2e09b4f_zeling_39674758421_1686640875//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2  RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡 stream CORTO","type":"archive","url":"https://www.twitch.tv/videos/1845060937","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":53928,"viewable":"public"},{"created_at":"2023-06-12T17:05:21Z","description":"","duration":"2h9m50s","id":"1844500473","language":"es","muted_segments":null,"published_at":"2023-06-12T17:05:21Z","stream_id":"39673161157","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/2da3c11328ff9ec12e1d_zeling_39673161157_1686589516//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844500473","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":17792,"viewable":"public"},{"created_at":"2023-06-12T06:37:33Z","description":"","duration":"7h56m10s","id":"1844222546","language":"es","muted_segments":null,"published_at":"2023-06-12T06:37:33Z","stream_id":"39671417045","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e4dae1bdae5f2a390746_zeling_39671417045_1686551846//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844222546","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":58274,"viewable":"public"}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"}}`)
	wantQuery := "first=1&period=week&type=archive&user_id=58753574"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(vodsJson)
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		c: sv.Client(),
	}
	vods, err := hx.Vods(&VODParams{
		BroadcasterID:  "58753574",
		Period:         Week,
		OnlyMostRecent: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(vods) != 1 {
		t.Fatalf("expected exactly 1 vod, got %d", len(vods))
	}
	if vods[0].VideoID != "1846472757" {
		t.Fatalf("expected vod to be 1846472757, got %s", vods[0].VideoID)
	}
}
