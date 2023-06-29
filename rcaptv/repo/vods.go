package repo

import (
	"database/sql"
	"strings"

	. "github.com/go-jet/jet/v2/postgres"

	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/helix"
)

type VodsParams struct {
	VideoIDs []string
	BcID string
	BcUsername string
	// If Extend > 0, Vods() will use `created_at` column of the last VOD
	// obtained and append the `Extend` number of VODs following the `created_at`
	// timestamp order. Extend won't work for Vods() querying by BcUsername, it's
	// mostly useful for queries with a single videoID
	Extend int
	First int
}

func Vods(db *sql.DB, p *VodsParams) ([]*helix.VOD, error) {
	if p.First == 0 {
		p.First = 1
	}
	if p.BcUsername != "" {
		return vodsByStreamer(db, p)
	}
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
	stmt = stmt.ORDER_BY(tbl.Vods.CreatedAt.DESC()).
		LIMIT(int64(p.First))

	var r []*helix.VOD
	if err := stmt.Query(db, &r); err != nil {
		return nil, err
	}

	if p.Extend > 0 && len(r) > 0{
		lastRow := r[len(r)-1]
		stmt2 := SELECT(
			tbl.Vods.AllColumns,
		).FROM(tbl.Vods).
		WHERE(
			tbl.Vods.CreatedAt.LT(TimestampT(lastRow.CreatedAt)).
		AND(
			tbl.Vods.BcID.EQ(String(lastRow.BroadcasterID)),
		)).ORDER_BY(tbl.Vods.CreatedAt.DESC()).LIMIT(int64(p.Extend))
		var r2 []*helix.VOD
		if err := stmt2.Query(db, &r2); err != nil {
			return nil, err
		}
		r = append(r, r2...)
	}
	return r, nil
}

func vodsByStreamer(db *sql.DB, p *VodsParams) (r []*helix.VOD, err error) {
	username := strings.ToLower(p.BcUsername)
	stmt := SELECT(
		tbl.Vods.AllColumns,
	).FROM(
		tbl.Vods.INNER_JOIN(
			tbl.TrackedChannels,
			tbl.TrackedChannels.BcID.EQ(tbl.Vods.BcID),
		),
	).WHERE(
		tbl.TrackedChannels.BcUsername.EQ(String(username)),
	).ORDER_BY(tbl.Vods.CreatedAt.DESC()).LIMIT(int64(p.First))

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
