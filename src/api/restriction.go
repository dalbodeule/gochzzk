package chzzk

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

// AddRestriction restricts a viewer channel on the authenticated channel.
func (c *Client) AddRestriction(ctx context.Context, targetChannelID string) error {
	return c.do(ctx, http.MethodPost, "/open/v1/restrict-channels", nil, struct {
		TargetChannelID string `json:"targetChannelId"`
	}{targetChannelID}, authAccessToken, nil)
}

// RemoveRestriction removes a viewer channel from the authenticated channel's restrictions.
func (c *Client) RemoveRestriction(ctx context.Context, targetChannelID string) error {
	return c.do(ctx, http.MethodDelete, "/open/v1/restrict-channels", nil, struct {
		TargetChannelID string `json:"targetChannelId"`
	}{targetChannelID}, authAccessToken, nil)
}

// GetRestrictions returns a cursor page of restricted viewer channels.
func (c *Client) GetRestrictions(ctx context.Context, size int, next string) (RestrictionPage, error) {
	q := url.Values{}
	request.IntParam(q, "size", size)
	if next != "" {
		q.Set("next", next)
	}
	var v RestrictionPage
	err := c.do(ctx, http.MethodGet, "/open/v1/restrict-channels", q, nil, authAccessToken, &v)
	return v, err
}
