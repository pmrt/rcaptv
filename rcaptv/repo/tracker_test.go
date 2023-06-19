package repo

import (
	"testing"

	"pedro.to/rcaptv/gen/tracker/public/model"
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
