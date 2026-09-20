package chzzk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const authorizationURL = "https://chzzk.naver.com/account-interlock"

// AuthorizationURL builds the URL where a user grants the application access.
// The redirect URI must exactly match the URI registered for the application.
func AuthorizationURL(clientID, redirectURI, state string) string {
	q := url.Values{}
	q.Set("clientId", clientID)
	q.Set("redirectUri", redirectURI)
	q.Set("state", state)
	return authorizationURL + "?" + q.Encode()
}

// Token is a CHZZK OAuth token response. RefreshToken is replaced on refresh.
type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    string `json:"expiresIn"`
	Scope        string `json:"scope,omitempty"`
}

// ExchangeCode exchanges an authorization code for an access/refresh token pair.
func (c *Client) ExchangeCode(ctx context.Context, code, state string) (Token, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return Token{}, fmt.Errorf("chzzk: client ID and client secret are required")
	}
	body := struct {
		GrantType    string `json:"grantType"`
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		State        string `json:"state"`
	}{"authorization_code", c.clientID, c.clientSecret, code, state}
	var token Token
	err := c.do(ctx, http.MethodPost, "/auth/v1/token", nil, body, authNone, &token)
	return token, err
}

// RefreshToken exchanges a refresh token for a new token pair.
// CHZZK refresh tokens are single-use; persist the returned RefreshToken.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (Token, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return Token{}, fmt.Errorf("chzzk: client ID and client secret are required")
	}
	body := struct {
		GrantType    string `json:"grantType"`
		RefreshToken string `json:"refreshToken"`
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	}{"refresh_token", refreshToken, c.clientID, c.clientSecret}
	var token Token
	err := c.do(ctx, http.MethodPost, "/auth/v1/token", nil, body, authNone, &token)
	return token, err
}

// RevokeToken invalidates tokens associated with the same client and user.
func (c *Client) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	if c.clientID == "" || c.clientSecret == "" {
		return fmt.Errorf("chzzk: client ID and client secret are required")
	}
	body := struct {
		ClientID      string `json:"clientId"`
		ClientSecret  string `json:"clientSecret"`
		Token         string `json:"token"`
		TokenTypeHint string `json:"tokenTypeHint,omitempty"`
	}{c.clientID, c.clientSecret, token, tokenTypeHint}
	return c.do(ctx, http.MethodPost, "/auth/v1/token/revoke", nil, body, authNone, nil)
}
