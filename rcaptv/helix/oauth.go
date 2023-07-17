package helix

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
)

type NotifyHandler func(*oauth2.Token) error

type notifyReuseTokenSource struct {
	new oauth2.TokenSource
	mu  sync.Mutex
	t   *oauth2.Token
	// NotifyHandler is called when the token is refreshed.
	f NotifyHandler
}

func (s *notifyReuseTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t.Valid() {
		return s.t, nil
	}
	t, err := s.new.Token()
	if err != nil {
		return nil, err
	}
	s.t = t
	return t, s.f(t)
}

type NotifyReuseTokenSourceOpts struct {
	OAuthConfig *oauth2.Config
	Notify      NotifyHandler
}

func NotifyReuseTokenSource(t *oauth2.Token, opts NotifyReuseTokenSourceOpts) oauth2.TokenSource {
	if opts.Notify == nil {
		opts.Notify = func(*oauth2.Token) error { return nil }
	}
	return &notifyReuseTokenSource{
		new: opts.OAuthConfig.TokenSource(context.Background(), t),
		t:   t,
		f:   opts.Notify,
	}
}
