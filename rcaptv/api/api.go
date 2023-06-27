package api

import (
	"database/sql"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/repo"
)

type API struct {
	db *sql.DB
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

func New(sto database.Storage) *API {
	return &API{
		db: sto.Conn(),
	}
}