package repo

import (
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"

	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/helix"
)

type VodsParams struct {
	VideoIDs []string
	BcID string
}

func Vods(db *sql.DB, p *VodsParams) (r []*helix.VOD, err error) {
	stmt := SELECT(
		tbl.Vods.AllColumns,
	).FROM(tbl.Vods)
	if l := len(p.VideoIDs); l > 0 {
		ids := make([]Expression, 0, l)
		for _, v := range p.VideoIDs {
			ids = append(ids, String(v))
		}
		stmt = stmt.WHERE(
			tbl.Vods.VideoID.IN(ids...),
		)
	}
	if bid := p.BcID; bid != "" {
		stmt = stmt.WHERE(
		tbl.Vods.BcID.EQ(String(bid)),
		)
	}

	if err = stmt.Query(db, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func UpsertVods(db *sql.DB, vods []*helix.VOD) error {
	stmt := tbl.Vods.INSERT(
		tbl.Vods.VideoID, tbl.Vods.BcID, tbl.Vods.StreamID, tbl.Vods.CreatedAt,
		tbl.Vods.PublishedAt, tbl.Vods.DurationSeconds, tbl.Vods.Lang, tbl.Vods.Title,
		tbl.Vods.ThumbnailURL, tbl.Vods.ViewCount,
	)
	for _, v := range vods {
		stmt.VALUES(
			v.VideoID, v.BroadcasterID, v.StreamID, v.CreatedAt,
			v.PublishedAt, v.Duration, v.Lang, v.Title,
			v.ThumbnailURL, v.ViewCount,
		)
	}
	stmt.ON_CONFLICT(tbl.Vods.VideoID).DO_UPDATE(
		SET(
			tbl.Vods.PublishedAt.SET(tbl.Vods.EXCLUDED.PublishedAt),
			tbl.Vods.DurationSeconds.SET(tbl.Vods.EXCLUDED.DurationSeconds),
			tbl.Vods.ThumbnailURL.SET(tbl.Vods.EXCLUDED.ThumbnailURL),
			tbl.Vods.Title.SET(tbl.Vods.EXCLUDED.Title),
			tbl.Vods.ViewCount.SET(tbl.Vods.EXCLUDED.ViewCount),
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
