package repo

import (
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"

	"pedro.to/rcaptv/gen/tracker/public/model"
	. "pedro.to/rcaptv/gen/tracker/public/table"
	"pedro.to/rcaptv/logger"
)

// Tracked fetches tracked channels
func Tracked(db *sql.DB) (r []*model.TrackedChannels, err error) {
	l := logger.New("", "query:Tracked")

	stmt := SELECT(
		TrackedChannels.AllColumns,
	).FROM(TrackedChannels)

	if err = stmt.Query(db, &r); err != nil {
		l.Error().Err(err).Msg("error while executing query")
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
	l := logger.New("", "query:LastVODByStreamer")

	stmt := SELECT(
		Vods.BcID,
		Vods.VideoID,
	).DISTINCT(Vods.BcID).
		FROM(
			Vods.INNER_JOIN(TrackedChannels, Vods.BcID.EQ(TrackedChannels.BcID)),
		).
		GROUP_BY(Vods.BcID, Vods.VideoID).
		ORDER_BY(Vods.BcID, Vods.CreatedAt.DESC())
	var r []*LastVOIDByStreamerResults = make([]*LastVOIDByStreamerResults, 0, 1000)
	if err := stmt.Query(db, &r); err != nil {
		l.Error().Err(err).Msg("error while executing query")
		return nil, err
	}
	return r, nil
}
