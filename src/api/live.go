package chzzk

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

// GetLives returns a cursor page of currently live broadcasts.
func (c *Client) GetLives(ctx context.Context, size int, next string) (LivesPage, error) {
	q := url.Values{}
	request.IntParam(q, "size", size)
	if next != "" {
		q.Set("next", next)
	}
	var v LivesPage
	err := c.do(ctx, http.MethodGet, "/open/v1/lives", q, nil, authClient, &v)
	return v, err
}

// GetStreamKey returns the authenticated channel's stream key.
func (c *Client) GetStreamKey(ctx context.Context) (StreamKey, error) {
	var v StreamKey
	err := c.do(ctx, http.MethodGet, "/open/v1/streams/key", nil, nil, authAccessToken, &v)
	return v, err
}

// GetLiveSetting returns the authenticated channel's broadcast settings.
func (c *Client) GetLiveSetting(ctx context.Context) (LiveSetting, error) {
	var v LiveSetting
	err := c.do(ctx, http.MethodGet, "/open/v1/lives/setting", nil, nil, authAccessToken, &v)
	return v, err
}

// UpdateLiveSetting partially updates the authenticated channel's broadcast settings.
func (c *Client) UpdateLiveSetting(ctx context.Context, patch LiveSettingPatch) error {
	return c.do(ctx, http.MethodPatch, "/open/v1/lives/setting", nil, patch, authAccessToken, nil)
}
