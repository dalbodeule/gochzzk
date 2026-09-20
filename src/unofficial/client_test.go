package unofficial

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestGetLiveStatusAndChatChannelID(t *testing.T) {
	transport := roundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/polling/v3/channels/channel/live-status" || r.URL.Query().Get("includePlayerRecommendContent") != "false" {
			t.Fatalf("unexpected live URL: %s", r.URL.String())
		}
		return response(`{"code":200,"message":null,"content":{"liveTitle":"live","status":"OPEN","chatChannelId":"chat"}}`), nil
	})
	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithBaseURLs("https://live.invalid", "https://comm.invalid"))
	status, err := client.GetLiveStatus(context.Background(), "channel")
	if err != nil || status.ChatChannelID != "chat" || status.LiveTitle != "live" {
		t.Fatalf("unexpected live status: %+v, err=%v", status, err)
	}
}

func TestGetFollowDate(t *testing.T) {
	transport := roundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/nng_main/v1/chats/chat/users/user/profile-card" || r.URL.Query().Get("chatType") != "STREAMING" {
			t.Fatalf("unexpected profile URL: %s", r.URL.String())
		}
		return response(`{"code":200,"message":null,"content":{"userIdHash":"user","streamingProperty":{"following":{"followDate":"2025-01-02T03:04:05Z"},"nicknameColor":{"colorCode":"#fff"}}}}`), nil
	})
	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithBaseURLs("https://live.invalid", "https://comm.invalid"))
	date, err := client.GetFollowDate(context.Background(), "chat", "user")
	if err != nil || date == nil || *date != "2025-01-02T03:04:05Z" {
		t.Fatalf("unexpected follow date: %v, err=%v", date, err)
	}
}
