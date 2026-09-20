// Package chzzk provides a typed client for the CHZZK Open API.
package chzzk

import (
	"context"
	"net/url"
	"strings"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

const defaultBaseURL = "https://openapi.chzzk.naver.com"

// HTTPDoer is implemented by *http.Client and makes Client easy to test.
type HTTPDoer = request.HTTPDoer

// Client is a CHZZK Open API client.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	accessToken  string
	userAgent    string
	httpClient   HTTPDoer
}
type Option func(*Client)

// WithHTTPClient replaces the HTTP transport used by Client.
func WithHTTPClient(h HTTPDoer) Option { return func(c *Client) { c.httpClient = h } }

// WithBaseURL overrides the API endpoint. It is primarily useful for tests.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(rawURL, "/") }
}

// WithClientCredentials sets credentials for client-authenticated endpoints.
func WithClientCredentials(id, secret string) Option {
	return func(c *Client) { c.clientID, c.clientSecret = id, secret }
}

// WithAccessToken sets a user-authorized OAuth access token.
func WithAccessToken(token string) Option { return func(c *Client) { c.accessToken = token } }

// WithUserAgent sets an optional User-Agent header on every request.
func WithUserAgent(userAgent string) Option { return func(c *Client) { c.userAgent = userAgent } }

// New creates a CHZZK API client.
func New(options ...Option) *Client {
	c := &Client{baseURL: defaultBaseURL, httpClient: request.NewHTTPClient()}
	for _, option := range options {
		option(c)
	}
	return c
}

// APIError is returned when CHZZK responds with a non-2xx status or error code.
type APIError = request.APIError
type authMode = request.AuthMode

const (
	authNone        = request.NoAuth
	authClient      = request.ClientAuth
	authAccessToken = request.AccessTokenAuth
)

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, auth authMode, out any) error {
	return request.Do(ctx, request.Config{BaseURL: c.baseURL, ClientID: c.clientID, ClientSecret: c.clientSecret, AccessToken: c.accessToken, UserAgent: c.userAgent, HTTPClient: c.httpClient}, method, path, query, body, auth, out)
}
