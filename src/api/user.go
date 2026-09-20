package chzzk

import (
	"context"
	"net/http"
)

// GetUser returns the channel belonging to the access-token user.
func (c *Client) GetUser(ctx context.Context) (User, error) {
	var v User
	err := c.do(ctx, http.MethodGet, "/open/v1/users/me", nil, nil, authAccessToken, &v)
	return v, err
}
