package repo

import (
	"testing"
	"time"

	"pedro.to/rcaptv/gen/tracker/public/model"
	"pedro.to/rcaptv/helix"
)

func TestTrackedChannels(t *testing.T) {
	rows, err := Tracked(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatal("expected tracked channels to return exactly 2 rows")
	}

	want := []*model.TrackedChannels{
		{
			BcID: "58753574",
		},
		{
			BcID: "90075649",
		},
	}
	for i, row := range rows {
		got := row.BcID
		want := want[i].BcID
		if got != want {
			t.Fatalf("unexpected tracked channel id, got %s, want %s", got, want)
		}
	}
}

func TestLastVODByStreamer(t *testing.T) {
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

func TestUpsertClips(t *testing.T) {
	vodOffset := 10
	clips := []*helix.Clip{
		{
			ClipID:           "clip1",
			BroadcasterID:    "58753574",
			VideoID:          "video1",
			CreatedAt:        "2023-01-01T10:00:00Z",
			CreatorID:        "creator1",
			CreatorName:      "John Doe",
			Title:            "Awesome Clip",
			GameID:           "game1",
			Lang:             "en",
			ThumbnailURL:     "https://example.com/thumbnail1.jpg",
			DurationSeconds:  10.5,
			ViewCount:        100,
			VODOffsetSeconds: nil,
		},
		{
			ClipID:           "clip2",
			BroadcasterID:    "58753574",
			VideoID:          "",
			CreatedAt:        "2023-02-15T15:30:00Z",
			CreatorID:        "creator2",
			CreatorName:      "Jane Smith",
			Title:            "Funny Clip",
			GameID:           "game2",
			Lang:             "es",
			ThumbnailURL:     "https://example.com/thumbnail2.jpg",
			DurationSeconds:  15.2,
			ViewCount:        250,
			VODOffsetSeconds: &vodOffset,
		},
	}
	UpsertClips(db, clips)
	got, err := Clips(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 clips, got %d", len(got))
	}
	if got[0].ViewCount != 100 {
		t.Fatalf("expected view count to be 100, got %d", got[0].ViewCount)
	}
	if got[1].ViewCount != 250 {
		t.Fatalf("expected view count to be 250, got %d", got[1].ViewCount)
	}

	vodOffset2 := 10
	clips2 := []*helix.Clip{
		{
			ClipID:           "clip1",
			BroadcasterID:    "58753574",
			VideoID:          "",
			CreatedAt:        "2023-01-01T10:00:00Z",
			CreatorID:        "creator1",
			CreatorName:      "John Doe",
			Title:            "Awesome Clip",
			GameID:           "game1",
			Lang:             "en",
			ThumbnailURL:     "https://example.com/thumbnail1.jpg",
			DurationSeconds:  10.5,
			ViewCount:        500,
			VODOffsetSeconds: &vodOffset2,
		},
		{
			ClipID:           "clip2",
			BroadcasterID:    "58753574",
			VideoID:          "video2",
			CreatedAt:        "2023-02-15T15:30:00Z",
			CreatorID:        "creator2",
			CreatorName:      "Jane Smith",
			Title:            "Funny Clip",
			GameID:           "game2",
			Lang:             "es",
			ThumbnailURL:     "https://example.com/thumbnail2.jpg",
			DurationSeconds:  15.2,
			ViewCount:        1000,
			VODOffsetSeconds: nil,
		},
	}
	UpsertClips(db, clips2)
	got, err = Clips(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 clips, got %d", len(got))
	}
	if got[0].ClipID != "clip1" {
		t.Fatalf("expected clip id to be 'clip1', got %s", got[0].ClipID)
	}
	if got[1].ClipID != "clip2" {
		t.Fatalf("expected clip id to be 'clip2', got %s", got[1].ClipID)
	}

	if *got[0].VODOffsetSeconds != vodOffset2 {
		t.Fatal("nil vodOffset => non-nil vodOffset merge should be: non-nil vodOffset")
	}
	if got[1].VODOffsetSeconds == nil {
		t.Fatal("non-nil vodOffset => nil vodOffset merge should be: non-nil vodOffset")
	}

	if got[0].VideoID != "video1" {
		t.Fatalf("non-empty video string => empty video string merge should be: non-empty video string")
	}
	if got[1].VideoID != "video2" {
		t.Fatalf("empty video string => non-empty video string merge should be: non-empty video string")
	}

	if got[0].ViewCount != 500 {
		t.Fatalf("expected view count to be 500, got %d", got[0].ViewCount)
	}
	if got[1].ViewCount != 1000 {
		t.Fatalf("expected view count to be 1000, got %d", got[1].ViewCount)
	}
}

func TestUpsertVods(t *testing.T) {
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
