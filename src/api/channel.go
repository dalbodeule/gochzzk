package chzzk

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

// GetChannels returns up to 20 public channels by ID.
func (c *Client) GetChannels(ctx context.Context, channelIDs []string) ([]Channel, error) {
	q := url.Values{}
	for _, id := range channelIDs {
		q.Add("channelIds", id)
	}
	var v []Channel
	err := c.do(ctx, http.MethodGet, "/open/v1/channels", q, nil, authClient, &v)
	return v, err
}

// GetStreamingRoles returns managers of the authenticated channel.
func (c *Client) GetStreamingRoles(ctx context.Context) ([]StreamingRole, error) {
	var v []StreamingRole
	err := c.do(ctx, http.MethodGet, "/open/v1/channels/streaming-roles", nil, nil, authAccessToken, &v)
	return v, err
}

// GetFollowers returns a page of followers. Page is zero-based and size is 1..50.
func (c *Client) GetFollowers(ctx context.Context, page, size int) ([]Follower, error) {
	q := url.Values{}
	request.IntParam(q, "page", page)
	request.IntParam(q, "size", size)
	var v []Follower
	err := c.do(ctx, http.MethodGet, "/open/v1/channels/followers", q, nil, authAccessToken, &v)
	return v, err
}

// GetSubscribers returns a page of subscribers. Sort is RECENT or LONGER.
func (c *Client) GetSubscribers(ctx context.Context, page, size int, sort string) ([]Subscriber, error) {
	q := url.Values{}
	request.IntParam(q, "page", page)
	request.IntParam(q, "size", size)
	if sort != "" {
		q.Set("sort", sort)
	}
	var v []Subscriber
	err := c.do(ctx, http.MethodGet, "/open/v1/channels/subscribers", q, nil, authAccessToken, &v)
	return v, err
}
