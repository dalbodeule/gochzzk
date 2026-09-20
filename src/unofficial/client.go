// Package unofficial wraps undocumented CHZZK web endpoints.
// These endpoints are not covered by the official Open API compatibility contract.
package unofficial

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

const (
	defaultLiveBaseURL = "https://api.chzzk.naver.com"
	defaultCommBaseURL = "https://comm-api.game.naver.com"
	maxResponseBytes   = 4 << 20
)

type HTTPDoer = request.HTTPDoer

type Client struct {
	httpClient  HTTPDoer
	liveBaseURL string
	commBaseURL string
	userAgent   string
}
type Option func(*Client)

func WithHTTPClient(client HTTPDoer) Option { return func(c *Client) { c.httpClient = client } }
func WithBaseURLs(live, comm string) Option {
	return func(c *Client) {
		c.liveBaseURL = strings.TrimRight(live, "/")
		c.commBaseURL = strings.TrimRight(comm, "/")
	}
}
func WithUserAgent(userAgent string) Option { return func(c *Client) { c.userAgent = userAgent } }

func New(options ...Option) *Client {
	c := &Client{httpClient: request.NewHTTPClient(), liveBaseURL: defaultLiveBaseURL, commBaseURL: defaultCommBaseURL}
	for _, option := range options {
		option(c)
	}
	return c
}

type APIError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("unofficial chzzk API error: HTTP %d, code %d: %s", e.StatusCode, e.Code, e.Message)
}

type envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Content T      `json:"content"`
}

func (c *Client) get(ctx context.Context, rawURL string, out any) (bool, error) {
	if c == nil || c.httpClient == nil {
		return false, errors.New("unofficial chzzk: HTTP client is nil")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return false, fmt.Errorf("create unofficial request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("send unofficial request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return false, fmt.Errorf("read unofficial response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return false, fmt.Errorf("unofficial chzzk: response exceeds %d bytes", maxResponseBytes)
	}
	var result envelope[json.RawMessage]
	if err := json.Unmarshal(data, &result); err != nil {
		return false, fmt.Errorf("decode unofficial response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || (result.Code != 0 && result.Code != 200) {
		return false, &APIError{StatusCode: resp.StatusCode, Code: result.Code, Message: result.Message}
	}
	if len(result.Content) == 0 || string(result.Content) == "null" {
		return false, nil
	}
	if err := json.Unmarshal(result.Content, out); err != nil {
		return false, fmt.Errorf("decode unofficial content: %w", err)
	}
	return true, nil
}

func escaped(value string) (string, error) {
	if value == "" {
		return "", errors.New("unofficial chzzk: identifier is required")
	}
	return url.PathEscape(value), nil
}
