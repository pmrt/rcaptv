package repo

import (
	"testing"
	"time"

	"pedro.to/rcaptv/helix"
)

func TestClipsUpsert(t *testing.T) {
	defer cleanupClips()

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
	if err := UpsertClips(db, clips); err != nil {
		t.Fatal(err)
	}
	got, err := Clips(db, nil)
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
	got, err = Clips(db, nil)
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

func TestClipsSelect(t *testing.T) {
	defer cleanupClips()

	bid := "58753574"
	vodOffset := 10
	clips := []*helix.Clip{
		{
			ClipID:           "clip1",
			BroadcasterID:    bid,
			VideoID:          "",
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
			BroadcasterID:    bid,
			VideoID:          "video1",
			CreatedAt:        "2023-02-15T10:01:00Z",
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
		{
			ClipID:           "clip3",
			BroadcasterID:    bid,
			VideoID:          "video2",
			CreatedAt:        "2023-02-15T10:30:00Z",
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
		{
			ClipID:           "clip4",
			BroadcasterID:    bid,
			VideoID:          "video3",
			CreatedAt:        "2023-02-15T10:50:00Z",
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
	if err := UpsertClips(db, clips); err != nil {
		t.Fatal(err)
	}
	got, err := Clips(db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 clips (all), got:%d", len(got))
	}

	got, err = Clips(db, &ClipsParams{
		BroadcasterID:   bid,
		ExcludeDangling: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 clips (with valid vod_offset), got:%d", len(got))
	}

	startedAt, err := time.Parse(time.RFC3339, "2023-02-15T10:29:00Z")
	if err != nil {
		t.Fatal(err)
	}
	got, err = Clips(db, &ClipsParams{
		BroadcasterID:   bid,
		StartedAt:       startedAt,
		ExcludeDangling: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 clips (after start_at), got:%d", len(got))
	}
	want := []string{"clip3", "clip4"}
	for i, clip := range got {
		if want[i] != clip.ClipID {
			t.Fatalf("unexpected clipID: got:'%s', want:'%s'", clip.ClipID, want)
		}
	}

	startedAt, err = time.Parse(time.RFC3339, "2023-02-15T10:29:00Z")
	if err != nil {
		t.Fatal(err)
	}
	endedAt, err := time.Parse(time.RFC3339, "2023-02-15T10:31:00Z")
	if err != nil {
		t.Fatal(err)
	}
	got, err = Clips(db, &ClipsParams{
		BroadcasterID:   bid,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		ExcludeDangling: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 clips (between start_at and ended_at), got:%d", len(got))
	}
	want = []string{"clip3"}
	for i, clip := range got {
		if want[i] != clip.ClipID {
			t.Fatalf("unexpected clipID: got:'%s', want:'%s'", clip.ClipID, want)
		}
	}

	startedAt, err = time.Parse(time.RFC3339, "2023-02-15T09:59:00Z")
	if err != nil {
		t.Fatal(err)
	}
	endedAt, err = time.Parse(time.RFC3339, "2023-02-15T10:05:00Z")
	if err != nil {
		t.Fatal(err)
	}
	got, err = Clips(db, &ClipsParams{
		BroadcasterID:   bid,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		ExcludeDangling: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 clips (after start_at, only with valid offset), got:%d", len(got))
	}
	want = []string{"clip2"}
	for i, clip := range got {
		if want[i] != clip.ClipID {
			t.Fatalf("unexpected clipID: got:'%s', want:'%s'", clip.ClipID, want)
		}
	}
}
