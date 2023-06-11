package repo

import (
	"log"
	"testing"

	"github.com/go-test/deep"
	"pedro.to/rcaptv/gen/tracker/public/model"
	. "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/utils"
)

func insertTrackedChannel(ch *model.TrackedChannels) {
	stmt := TrackedChannels.INSERT(
		TrackedChannels.AllColumns,
	).MODEL(ch)

	_, err := stmt.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
}

func TestTrackedChannels(t *testing.T) {
	insertTrackedChannel(&model.TrackedChannels{
		BcID:          "36138196",
		BcDisplayName: "alexelcapo",
		BcUsername:    "alexelcapo",
		BcType:        model.Broadcastertype_Partner,
		PpURL:         utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
		OfflinePpURL:  utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
	})

	rows, err := Tracked(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatal("expected tracked channels to return exactly 1 row")
	}

	want := &model.TrackedChannels{
		BcID:          "36138196",
		BcDisplayName: "alexelcapo",
		BcUsername:    "alexelcapo",
		BcType:        model.Broadcastertype_Partner,
		PpURL:         utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
		OfflinePpURL:  utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
	}
	if diff := deep.Equal(rows[0], want); diff != nil {
		t.Fatal(diff)
	}
}
