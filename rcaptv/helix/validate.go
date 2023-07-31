package helix

import (
	"context"
	"fmt"
	"net/http"
)

const TwitchValidateEndpoint = "https://id.twitch.tv/oauth2/validate"

type ValidateTokenResponse struct {
	ClientID     string   `json:"client_id"`
	Login        string   `json:"login"`
	Scopes       []string `json:"scopes"`
	TwitchUserID string   `json:"user_id"`
	ExpiresIn    int      `json:"expires_in"`
}

type ValidTokenParams struct {
	AccessToken string
	Context     context.Context
}

// ValidToken checks whether the provided access token is valid with the
// Twitch API
func (hx *Helix) ValidToken(p ValidTokenParams) bool {
	if p.Context == nil {
		p.Context = context.Background()
	}
	req, err := http.NewRequest("GET", hx.ValidateEndpoint(), nil)
	if err != nil {
		return false
	}
	ctx := ContextWithCustomQueryOpts(p.Context, &CustomQueryOpts{
		UseClientID: false,
	})
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", p.AccessToken))

	resp, err := hx.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}
