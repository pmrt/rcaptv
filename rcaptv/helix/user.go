package helix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type UserParams struct {
	UserID string
	Login  string

	Context context.Context
}

type User struct {
	Id              string           `json:"id"`
	Login           string           `json:"login"`
	DisplayName     string           `json:"display_name"`
	Type            string           `json:"type"`
	BroadcasterType string           `json:"broadcaster_type"`
	Description     string           `json:"description"`
	ProfileImageURL string           `json:"profile_image_url"`
	OfflineImageURL string           `json:"offline_image_url"`
	ViewCount       int              `json:"view_count"`
	Email           string           `json:"email"`
	CreatedAt       RFC3339Timestamp `json:"created_at"`
}

type UserResponse struct {
	Data []User
}

func (hx *Helix) User(p *UserParams) (*UserResponse, error) {
	params := url.Values{}

	if p.UserID != "" {
		params.Add("id", p.UserID)
	}
	if p.Login != "" {
		params.Add("login", p.Login)
	}
	if p.Context == nil {
		p.Context = context.Background()
	}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/users?%s", hx.APIUrl(), params.Encode()),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(p.Context)

	resp, err := hx.Do(req)
	if err != nil {
		return nil, err
	}
	var usr *UserResponse
	if err := json.Unmarshal(resp.Body, &usr); err != nil {
		return nil, err
	}
	return usr, nil
}
