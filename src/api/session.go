package chzzk

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

// CreateClientSession requests a Socket.IO URL using client authentication.
func (c *Client) CreateClientSession(ctx context.Context) (SessionURL, error) {
	return c.createSession(ctx, "/open/v1/sessions/auth/client", authClient)
}

// CreateUserSession requests a Socket.IO URL using the access token.
func (c *Client) CreateUserSession(ctx context.Context) (SessionURL, error) {
	return c.createSession(ctx, "/open/v1/sessions/auth", authAccessToken)
}

func (c *Client) createSession(ctx context.Context, path string, auth authMode) (SessionURL, error) {
	var v SessionURL
	err := c.do(ctx, http.MethodGet, path, nil, nil, auth, &v)
	return v, err
}

// ListClientSessions lists sessions created with client authentication.
func (c *Client) ListClientSessions(ctx context.Context, page string, size int) (SessionPage, error) {
	return c.listSessions(ctx, "/open/v1/sessions/client", page, size, authClient)
}

// ListUserSessions lists sessions created with the access token.
func (c *Client) ListUserSessions(ctx context.Context, page string, size int) (SessionPage, error) {
	return c.listSessions(ctx, "/open/v1/sessions", page, size, authAccessToken)
}

func (c *Client) listSessions(ctx context.Context, path, page string, size int, auth authMode) (SessionPage, error) {
	q := url.Values{}
	if page != "" {
		q.Set("page", page)
	}
	request.IntParam(q, "size", size)
	var v SessionPage
	err := c.do(ctx, http.MethodGet, path, q, nil, auth, &v)
	return v, err
}

// SubscribeChat subscribes a session to chat events.
func (c *Client) SubscribeChat(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/subscribe/chat", sessionKey)
}

// UnsubscribeChat unsubscribes a session from chat events.
func (c *Client) UnsubscribeChat(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/unsubscribe/chat", sessionKey)
}

// SubscribeDonation subscribes a session to donation events.
func (c *Client) SubscribeDonation(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/subscribe/donation", sessionKey)
}

// UnsubscribeDonation unsubscribes a session from donation events.
func (c *Client) UnsubscribeDonation(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/unsubscribe/donation", sessionKey)
}

// SubscribeSubscription subscribes a session to subscription events.
func (c *Client) SubscribeSubscription(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/subscribe/subscription", sessionKey)
}

// UnsubscribeSubscription unsubscribes a session from subscription events.
func (c *Client) UnsubscribeSubscription(ctx context.Context, sessionKey string) error {
	return c.sessionEvent(ctx, "/open/v1/sessions/events/unsubscribe/subscription", sessionKey)
}

func (c *Client) sessionEvent(ctx context.Context, path, sessionKey string) error {
	q := url.Values{"sessionKey": []string{sessionKey}}
	return c.do(ctx, http.MethodPost, path, q, nil, authAccessToken, nil)
}
