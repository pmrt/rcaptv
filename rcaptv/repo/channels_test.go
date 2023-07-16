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
