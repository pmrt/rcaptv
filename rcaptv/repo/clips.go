package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"

	tbl "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/helix"
)

type ClipsParams struct {
	BroadcasterID string
	StartedAt     time.Time
	EndedAt       time.Time

	// ExcludeDangling excludes clips that have no connection with vods
	// (determined by vod_offset)
	ExcludeDangling bool

	Context context.Context
}

func Clips(db *sql.DB, p *ClipsParams) (r []*helix.Clip, err error) {
	stmt := SELECT(
		tbl.Clips.AllColumns,
	).FROM(tbl.Clips)

	var ctx context.Context
	if p != nil {
		if p.Context != nil {
			ctx = p.Context
		} else {
			ctx = context.Background()
		}
		if p.BroadcasterID == "" {
			return nil, errors.New("empty broadcaster id")
		}
		where := tbl.Clips.BcID.EQ(String(p.BroadcasterID))
		if !p.StartedAt.IsZero() {
			where = where.AND(tbl.Clips.CreatedAt.GT(TimestampT(p.StartedAt)))
		}
		if !p.EndedAt.IsZero() {
			where = where.AND(tbl.Clips.CreatedAt.LT(TimestampT(p.EndedAt)))
		}
		if p.ExcludeDangling {
			where = where.AND(tbl.Clips.VodOffset.IS_NOT_NULL())
		}
		stmt = stmt.WHERE(where)
	} else {
		ctx = context.Background()
	}

	if err = stmt.QueryContext(ctx, db, &r); err != nil {
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
			c.ClipID, c.BroadcasterID, c.VideoID, c.CreatedAt, c.CreatorID,
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
	return ErrNoRowsAffected
}
