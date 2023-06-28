package repo

import (
	"testing"
	"time"

	"pedro.to/rcaptv/helix"
)

func TestLastVODByStreamer(t *testing.T) {
	t.Parallel()
	rows, err := LastVODByStreamer(db)
	if err != nil {
		t.Fatal(err)
	}

	wantRows := []struct {
		bid string
		vid string
	}{
		{bid: "58753574", vid: "1849520474"},
		{bid: "90075649", vid: "1847800606"},
	}

	for i, row := range rows {
		got := row.BroadcasterID
		want := wantRows[i].bid
		if got != want {
			t.Fatalf("unexpected bid, got %s, want %s", got, want)
		}

		got = row.VodID
		want = wantRows[i].vid
		if got != want {
			t.Fatalf("unexpected vid, got %s, want %s", got, want)
		}
	}
}

func TestVodsByStreamer(t *testing.T) {
	vods, err := Vods(db, &VodsParams{
		BcUsername: "IlloJuan", // display name should work too
	})
	if err != nil {
		t.Fatal(err)
	}
	wantIds := []string{"1847800606", "1846954069", "1846151378", "1845269425"}
	for i, vod := range vods {
		got := vod.VideoID
		want := wantIds[i]
		if got != want {
			t.Fatalf("unexpected vod, got %s, want %s", got, want)
		}
	}
}

func TestUpsertVods(t *testing.T) {
	t.Parallel()
	ts1, err := time.Parse(time.RFC3339, "2023-06-14T23:21:38Z")
	if err != nil {
		t.Fatal(err)
	}
	ts2, err := time.Parse(time.RFC3339, "2023-06-14T07:21:23Z")
	if err != nil {
		t.Fatal(err)
	}
	vods := []*helix.VOD{
		{
			VideoID:        "18464727000",
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
			VideoID:        "1845909001",
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
	}
	UpsertVods(db, vods)
	got, err := Vods(db, &VodsParams{
		VideoIDs: []string{"18464727000", "1845909001"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 vods, got %d", len(got))
	}
	if got[0].Duration != 4000 {
		t.Fatalf("expected duration to be 4000, got %d", got[0].Duration)
	}
	if got[1].Duration != 21350 {
		t.Fatalf("expected duration to be 21350, got %d", got[1].Duration)
	}
	if got[1].Title != "☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡" {
		t.Fatalf("expected title to be 'test', got %s", got[1].Title)
	}
	if got[0].ViewCount != 7865 {
		t.Fatalf("expected view count to be 10000, got %d", got[0].ViewCount)
	}

	vods2 := []*helix.VOD{
		{
			VideoID:        "18464727000",
			BroadcasterID:  "58753574",
			StreamID:       "46935025356",
			CreatedAt:      ts1,
			PublishedAt:    ts1,
			DurationString: "1h6m40s",
			Lang:           "es",
			Title:          "☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      10000,
			Duration:       4810,
		},
		{
			VideoID:        "1845909001",
			BroadcasterID:  "58753574",
			StreamID:       "46932511084",
			CreatedAt:      ts2,
			PublishedAt:    ts2,
			DurationString: "5h55m50s",
			Lang:           "es",
			Title:          "test",
			ThumbnailURL:   "https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg",
			ViewCount:      39695,
			Duration:       30000,
		},
	}
	UpsertVods(db, vods2)
	got, err = Vods(db, &VodsParams{
		VideoIDs: []string{"18464727000", "1845909001"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 vods, got %d", len(got))
	}
	if got[0].Duration != 4810 {
		t.Fatalf("expected duration to be 4000, got %d", got[0].Duration)
	}
	if got[1].Duration != 30000 {
		t.Fatalf("expected duration to be 21350, got %d", got[1].Duration)
	}
	if got[1].Title != "test" {
		t.Fatalf("expected title to be 'test', got %s", got[1].Title)
	}
	if got[0].ViewCount != 10000 {
		t.Fatalf("expected view count to be 10000, got %d", got[0].ViewCount)
	}
}
