package api

import (
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type CredentialCookie string

func (c *CredentialCookie) Parse() (int64, *oauth2.Token) {
	s := strings.Split(string(*c), ":")
	if len(s) != 4 {
		return 0, nil
	}
	id, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return 0, nil
	}
	exp, err := time.Parse(time.RFC3339, s[3])
	if err != nil {
		return 0, nil
	}
	return id, &oauth2.Token{
		AccessToken:  s[1],
		RefreshToken: s[2],
		Expiry:       exp,
		TokenType:    "Bearer",
	}
}

func (c *CredentialCookie) Set(userID int64, t *oauth2.Token) string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatInt(userID, 10))
	sb.WriteString(":")
	sb.WriteString(t.AccessToken)
	sb.WriteString(":")
	sb.WriteString(t.RefreshToken)
	sb.WriteString(":")
	sb.WriteString(t.Expiry.Format(time.RFC3339))
	*c = CredentialCookie(sb.String())
	return string(*c)
}
