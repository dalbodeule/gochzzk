package chzzk

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(body string, status int) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestGetLivesUsesClientAuthentication(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.URL.Path != "/open/v1/lives" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Client-Id"); got != "id" {
			t.Errorf("Client-Id = %q", got)
		}
		if got := r.Header.Get("Client-Secret"); got != "secret" {
			t.Errorf("Client-Secret = %q", got)
		}
		if got := r.URL.Query().Get("size"); got != "10" {
			t.Errorf("size = %q", got)
		}
		return response(`{"code":200,"message":null,"content":{"data":[{"liveId":7,"channelName":"테스트"}],"page":{"next":"cursor"}}}`, http.StatusOK), nil
	})

	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithClientCredentials("id", "secret"))
	page, err := client.GetLives(context.Background(), 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.Data[0].LiveID != 7 || page.Page.Next != "cursor" {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestGetUserUsesBearerToken(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("Authorization = %q", got)
		}
		return response(`{"code":200,"message":null,"content":{"channelId":"abc","channelName":"name"}}`, http.StatusOK), nil
	})

	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token"))
	user, err := client.GetUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.ChannelID != "abc" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestAPIError(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(`{"code":403,"message":"FORBIDDEN"}`, http.StatusForbidden), nil
	})

	_, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).GetUser(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Code != 403 || apiErr.Message != "FORBIDDEN" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

func TestUpdateLiveSettingOmitsNilFields(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"categoryId":"game"}` {
			t.Errorf("body = %s", body)
		}
		return response(`{"code":200,"message":null,"content":null}`, http.StatusOK), nil
	})

	id := "game"
	err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).UpdateLiveSetting(context.Background(), LiveSettingPatch{CategoryID: &id})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMissingCredentials(t *testing.T) {
	client := New()
	_, err := client.GetUser(context.Background())
	if err == nil || !strings.Contains(err.Error(), "access token") {
		t.Fatalf("unexpected error: %v", err)
	}
}
