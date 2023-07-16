package repo

import (
	"testing"

	"pedro.to/rcaptv/helix"
)

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
