// Package request contains implementation details shared by the public API packages.
package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxResponseBytes = 4 << 20
	defaultTimeout   = 30 * time.Second
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
type AuthMode uint8

const (
	NoAuth AuthMode = iota
	ClientAuth
	AccessTokenAuth
)

type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AccessToken  string
	UserAgent    string
	HTTPClient   HTTPDoer
}

// NewHTTPClient returns the hardened default client used by public clients.
// Redirects are returned to the caller instead of followed so credentials are
// never copied to a redirect target.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

type APIError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("chzzk API error: HTTP %d (code %d)", e.StatusCode, e.Code)
	}
	return fmt.Sprintf("chzzk API error: HTTP %d (code %d): %s", e.StatusCode, e.Code, e.Message)
}

type envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Content T      `json:"content"`
}

// Do executes a CHZZK API request and decodes its common response envelope.
func Do(ctx context.Context, config Config, method, path string, query url.Values, body any, auth AuthMode, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	u := strings.TrimRight(config.BaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if config.UserAgent != "" {
		req.Header.Set("User-Agent", config.UserAgent)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	switch auth {
	case AccessTokenAuth:
		if config.AccessToken == "" {
			return errors.New("chzzk: access token is required")
		}
		req.Header.Set("Authorization", "Bearer "+config.AccessToken)
	case ClientAuth:
		if config.ClientID == "" || config.ClientSecret == "" {
			return errors.New("chzzk: client ID and client secret are required")
		}
		req.Header.Set("Client-Id", config.ClientID)
		req.Header.Set("Client-Secret", config.ClientSecret)
	}
	if config.HTTPClient == nil {
		return errors.New("chzzk: HTTP client is nil")
	}
	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return fmt.Errorf("chzzk: response exceeds %d bytes", maxResponseBytes)
	}
	var response envelope[json.RawMessage]
	if err := json.Unmarshal(data, &response); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &APIError{StatusCode: resp.StatusCode, Message: string(data)}
		}
		return fmt.Errorf("decode response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || response.Code < 200 || response.Code >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Code: response.Code, Message: response.Message}
	}
	if out != nil && len(response.Content) > 0 && string(response.Content) != "null" {
		if err := json.Unmarshal(response.Content, out); err != nil {
			return fmt.Errorf("decode content: %w", err)
		}
	}
	return nil
}
