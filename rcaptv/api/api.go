package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

type API struct {
	db *sql.DB
	hx *helix.Helix // note: this is temporal. We should use user tokens
}

type APIResponse[T any] struct {
	Data T `json:"data"`
	Errors []string `json:"errors"`
}

func NewResponse[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{
		Data: data,
		Errors: make([]string, 0, 2),
	}
}

type VodsResponse struct {
	Vods []*helix.VOD `json:"vods"`
}
func (a *API) Vods(c *fiber.Ctx) error {
	resp := NewResponse(&VodsResponse{
		Vods: make([]*helix.VOD, 0, 5),
	})

	bid := c.Query("bid")
	if bid == "" {
		resp.Errors = append(resp.Errors, "Missing bid")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	vods, err := repo.Vods(a.db, &repo.VodsParams{
		BcID: bid,
	})
	if err != nil {
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.Status(http.StatusInternalServerError).JSON(resp)
	}

	resp.Data.Vods = append(resp.Data.Vods, vods...)
	return c.Status(http.StatusOK).JSON(resp)
}

type ClipsResponse struct {
	Clips []*helix.Clip `json:"clips"`
}
// Clips
// - `bid` string Broacaster ID
// - `started_at` string Start range time of creation of the clip in RFC3339
// - `ended_at` string End range time of creation of the clip in RFC3339
//
// Note: Twitch API does not provide a way to fetch clips by video_id.
// Alternative is to ask for bid+start+end of the stream. This may leave out
// some interesting clips created after the stream and include clips from other
// vods. One potential solution is to merge with tracker clips in server and
// filter by vod id in the client
func (a *API) Clips(c *fiber.Ctx) error {
	resp := NewResponse(&ClipsResponse{
		Clips: make([]*helix.Clip, 0, 10),
	})

	bid := c.Query("bid")
	if bid == "" {
		resp.Errors = append(resp.Errors, "Missing bid")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	var err error
	started := time.Time{}
	if s := c.Query("started_at"); s != "" {
		started, err = time.Parse(time.RFC3339, s)
		if err != nil {
			resp.Errors = append(resp.Errors, "Invalid 'started_at'")
			return c.Status(http.StatusBadRequest).JSON(resp)
		}
	}
	ended := time.Time{}
	if s := c.Query("ended_at"); s != "" {
		ended, err = time.Parse(time.RFC3339, s)
		if err != nil {
			resp.Errors = append(resp.Errors, "Invalid 'ended_at'")
			return c.Status(http.StatusBadRequest).JSON(resp)
		}
	}

	clips, err := a.hx.DeepClips(&helix.DeepClipsParams{
		ClipsParams: &helix.ClipsParams{
			BroadcasterID: bid,
			StartedAt: started,
			EndedAt: ended,
			StopViewsThreshold: cfg.ClipViewThreshold,
			ViewsThresholdWindowSize: cfg.ClipViewWindowSize,
		},
		MaxDeepLvl: cfg.ClipTrackingMaxDeepLevel,
	})
	if err != nil {
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.Status(http.StatusInternalServerError).JSON(resp)
	}
	resp.Data.Clips = append(resp.Data.Clips, clips...)
	return c.Status(http.StatusOK).JSON(resp)
}

func New(sto database.Storage, hx *helix.Helix) *API {
	return &API{
		db: sto.Conn(),
		hx: hx,
	}
}