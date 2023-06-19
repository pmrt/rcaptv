package repo

import (
	"database/sql"
	"errors"

	. "github.com/go-jet/jet/v2/postgres"

	"pedro.to/rcaptv/gen/tracker/public/model"
	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/helix"
)

var ErrNoRowsInserted = errors.New("no rows inserted")

// Tracked fetches tracked channels
func Tracked(db *sql.DB) (r []*model.TrackedChannels, err error) {
	stmt := SELECT(
		tbl.TrackedChannels.AllColumns,
	).FROM(tbl.TrackedChannels)

	if err = stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}

/*
select distinct on (vods.bc_id) vods.video_id, vods.bc_id
from vods
inner join tracked_channels ON vods.bc_id = tracked_channels.bc_id
group by vods.bc_id, vods.video_id
order by vods.bc_id, vods.created_at DESC;
*/

type LastVOIDByStreamerResults struct {
	BroadcasterID string `alias:"vods.bc_id"`
	VodID         string `alias:"vods.video_id"`
}

func LastVODByStreamer(db *sql.DB) ([]*LastVOIDByStreamerResults, error) {
	stmt := SELECT(
		tbl.Vods.BcID,
		tbl.Vods.VideoID,
	).DISTINCT(tbl.Vods.BcID).
		FROM(
			tbl.Vods.INNER_JOIN(tbl.TrackedChannels, tbl.Vods.BcID.EQ(tbl.TrackedChannels.BcID)),
		).
		GROUP_BY(tbl.Vods.BcID, tbl.Vods.VideoID).
		ORDER_BY(tbl.Vods.BcID, tbl.Vods.CreatedAt.DESC())
	var r []*LastVOIDByStreamerResults = make([]*LastVOIDByStreamerResults, 0, 1000)
	if err := stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func Clips(db *sql.DB) (r []*helix.Clip, err error) {
	stmt := SELECT(
		tbl.Clips.AllColumns,
	).FROM(tbl.Clips)

	if err = stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func UpsertClips(db *sql.DB, clips []*helix.Clip) error {
	stmt := tbl.Clips.INSERT(
		tbl.Clips.ClipID, tbl.Clips.BcID, tbl.Clips.VideoID, tbl.Clips.CreatedAt, tbl.Clips.CreatorID,
		tbl.Clips.CreatorName, tbl.Clips.Title, tbl.Clips.GameID, tbl.Clips.Lang, tbl.Clips.ThumbnailURL,
		tbl.Clips.DurationSeconds, tbl.Clips.ViewCount, tbl.Clips.VodOffset,
	)
	for _, c := range clips {
		stmt.VALUES(
			c.ClipID, c.BroadCasterID, c.VideoID, c.CreatedAt, c.CreatorID,
			c.CreatorName, c.Title, c.GameID, c.Lang, c.ThumbnailURL,
			c.DurationSeconds, c.ViewCount, c.VODOffsetSeconds,
		)
	}
	stmt.ON_CONFLICT(tbl.Clips.ClipID).DO_UPDATE(
		SET(
			tbl.Clips.Title.SET(tbl.Clips.EXCLUDED.Title),
			tbl.Clips.ViewCount.SET(tbl.Clips.EXCLUDED.ViewCount),
			tbl.Clips.VideoID.SET(
				StringExp(COALESCE(NULLIF(tbl.Clips.EXCLUDED.VideoID, String("")), tbl.Clips.VideoID)),
			),
			tbl.Clips.VodOffset.SET(
				IntExp(COALESCE(tbl.Clips.EXCLUDED.VodOffset, tbl.Clips.VodOffset)),
			),
		))
	res, err := stmt.Exec(db)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return ErrNoRowsInserted
}
