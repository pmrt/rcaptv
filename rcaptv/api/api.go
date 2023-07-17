package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
	"pedro.to/rcaptv/utils"
)

var (
	scopes = []string{"user:read:email"}
)

type APIOpts struct {
	Storage database.Storage
	Helix   *helix.Helix
	// Max difference between start/end for clips.
	ClipsMaxPeriodDiffHours int

	ClientID, ClientSecret string
	HelixAPIUrl            string
	HelixEventsubEndpoint  string
	AuthRedirectURL        string
}

type API struct {
	db          *sql.DB
	hx          *helix.Helix // note: this is temporal. We should use user tokens
	OAuthConfig *oauth2.Config

	clipsMaxPeriodDiffHours int
}

type APIResponse[T any] struct {
	Data   T        `json:"data"`
	Errors []string `json:"errors"`
}

func NewResponse[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{
		Data:   data,
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

	username := c.Query("username")
	vid := c.Query("vid")
	after := c.Query("after")

	var vids = make([]string, 0, 1)
	if username == "" {
		if vid == "" {
			if after == "" {
				resp.Errors = append(resp.Errors, "Missing username, after or vid")
				return c.Status(http.StatusBadRequest).JSON(resp)
			}
		} else {
			vids = append(vids, vid)
		}
	}
	ext, err := strconv.Atoi(c.Query("extend", "0"))
	if err != nil {
		resp.Errors = append(resp.Errors, "Bad extend value")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	vods, err := repo.Vods(a.db, &repo.VodsParams{
		VideoIDs:   vids,
		BcUsername: username,
		Extend:     utils.Min(ext, 5),
		After:      after,
	})
	if err != nil {
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.Status(http.StatusInternalServerError).JSON(resp)
	}
	if len(vods) == 0 {
		if username != "" {
			// TODO - this is a good thing to log, so we know which channels are our users interested in
			resp.Errors = append(resp.Errors, fmt.Sprintf("Username '%s' not found. The channel may not be tracked by us.", username))
		} else if vid != "" {
			resp.Errors = append(resp.Errors, fmt.Sprintf("VOD '%s' not found", vid))
		} else if after != "" {
			resp.Errors = append(resp.Errors, fmt.Sprintf("No VODS after '%s' found", after))
		}
		return c.Status(http.StatusNotFound).JSON(resp)
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
	started_at := c.Query("started_at")
	if started_at == "" {
		resp.Errors = append(resp.Errors, "Missing started_at")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	ended_at := c.Query("ended_at")
	if ended_at == "" {
		resp.Errors = append(resp.Errors, "Missing ended_at")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	started, err := time.Parse(time.RFC3339, started_at)
	if err != nil {
		resp.Errors = append(resp.Errors, "Invalid 'started_at'")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}
	ended, err := time.Parse(time.RFC3339, ended_at)
	if err != nil {
		resp.Errors = append(resp.Errors, "Invalid 'ended_at'")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	if ended.Sub(started) > time.Duration(a.clipsMaxPeriodDiffHours)*time.Hour {
		resp.Errors = append(resp.Errors, "period between 'started_at' and 'ended_at' is too large")
		return c.Status(http.StatusBadRequest).JSON(resp)
	}

	// check err and delete cookies if 401
	// test hx.clips with context
	// test hx.vods with context
	// /login, /logout, /auth/redirect <----- guardar en cookie id, at, rt
	// /validate token
	clips, err := a.hx.DeepClips(&helix.DeepClipsParams{
		ClipsParams: &helix.ClipsParams{
			BroadcasterID:            bid,
			StartedAt:                started,
			EndedAt:                  ended,
			StopViewsThreshold:       cfg.ClipViewThreshold,
			ViewsThresholdWindowSize: cfg.ClipViewWindowSize,
			Context:                  c.UserContext(),
		},
		MaxDeepLvl: cfg.ClipTrackingMaxDeepLevel,
	})
	c, err = a.checkErr(c, err)
	if err != nil {
		if errors.Is(err, helix.ErrItemsEmpty) {
			resp.Errors = append(resp.Errors, fmt.Sprintf("No clips found for the provided streamer (bid:'%s'). Are clips enabled for this streamer?", bid))
			return c.JSON(resp)
		}
		resp.Errors = append(resp.Errors, "Unexpected error")
		return c.JSON(resp)
	}
	resp.Data.Clips = append(resp.Data.Clips, clips...)
	return c.Status(http.StatusOK).JSON(resp)
}

func (a *API) checkErr(c *fiber.Ctx, err error) (*fiber.Ctx, error) {
	if err == nil {
		return c, nil
	}

	if errors.Is(err, helix.ErrUnauthorized) {
		a.clearAuthCookies(c)
		return c.Status(http.StatusUnauthorized), helix.ErrUnauthorized
	}

	if errors.Is(err, helix.ErrItemsEmpty) {
		return c.Status(http.StatusNotFound), errors.New("no items found")
	}

	return c.Status(http.StatusInternalServerError), fmt.Errorf("unexpected error: %w", err)
}

func New(opts APIOpts) *API {
	if opts.ClipsMaxPeriodDiffHours == 0 {
		opts.ClipsMaxPeriodDiffHours = 24 * 7
	}
	// o2cfg := &oauth2.Config{
	// 	ClientID:     opts.ClientID,
	// 	ClientSecret: opts.ClientSecret,
	// 	Scopes:       scopes,
	// 	Endpoint:     twitch.Endpoint,
	// 	RedirectURL:  fmt.Sprintf("%s:%s%s%s", cfg.BaseURL, cfg.APIPort, cfg.AuthEndpoint, cfg.AuthRedirectEndpoint),
	// }
	api := &API{
		db: opts.Storage.Conn(),
		hx: helix.NewWithUserTokens(&helix.HelixOpts{
			Creds: helix.ClientCreds{
				ClientID:     opts.ClientID,
				ClientSecret: opts.ClientSecret,
			},
			APIUrl:           opts.HelixAPIUrl,
			EventsubEndpoint: cfg.EventSubEndpoint,
		}),
		clipsMaxPeriodDiffHours: opts.ClipsMaxPeriodDiffHours,
	}
	return api
}
