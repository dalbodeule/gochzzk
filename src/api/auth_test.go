package chzzk

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExchangeCode(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if r.URL.Path != "/auth/v1/token" || !strings.Contains(string(body), `"grantType":"authorization_code"`) || !strings.Contains(string(body), `"code":"code"`) {
			t.Fatalf("unexpected token request: %s %s", r.URL.Path, body)
		}
		return response(`{"code":200,"message":null,"content":{"accessToken":"access","refreshToken":"refresh","tokenType":"Bearer","expiresIn":"86400"}}`, http.StatusOK), nil
	})
	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithClientCredentials("id", "secret"))
	token, err := client.ExchangeCode(context.Background(), "code", "state")
	if err != nil || token.AccessToken != "access" || token.RefreshToken != "refresh" {
		t.Fatalf("unexpected token: %+v, err=%v", token, err)
	}
}
