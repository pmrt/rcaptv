package helix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (hx *Helix) CreateEventsubSubscription(sub *Subscription) error {
	b := struct {
		Type      string     `json:"type"`
		Version   string     `json:"version"`
		Condition *Condition `json:"condition"`
		Transport *Transport `json:"transport"`
	}{
		Type:      sub.Type,
		Version:   sub.Version,
		Condition: sub.Condition,
		Transport: sub.Transport,
	}

	buf := bytes.NewBuffer(make([]byte, 0, EstimatedSubscriptionJSONSize))
	if err := json.NewEncoder(buf).Encode(b); err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		hx.opts.APIUrl+hx.opts.EventsubEndpoint+"/subscriptions",
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := hx.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Expected 200 response, got " + fmt.Sprint(resp.StatusCode))
	}
	return nil
}
