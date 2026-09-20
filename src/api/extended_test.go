package chzzk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAuthorizationURL(t *testing.T) {
	got := AuthorizationURL("id", "https://localhost/callback", "state")
	if !strings.Contains(got, "clientId=id") || !strings.Contains(got, "redirectUri=https%3A%2F%2Flocalhost%2Fcallback") || !strings.Contains(got, "state=state") {
		t.Fatalf("unexpected authorization URL: %s", got)
	}
}

func TestGetChannelLiveStatus(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(`{"code":200,"message":null,"content":{"data":[{"liveId":1,"channelId":"target","liveThumbnailImageUrl":"https://image"}],"page":{"next":""}}}`, http.StatusOK), nil
	})
	status, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithClientCredentials("id", "secret")).GetChannelLiveStatus(context.Background(), "target")
	if err != nil {
		t.Fatal(err)
	}
	if !status.IsLive || status.Live == nil || status.ThumbnailURL != "https://image" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestRestrictionRequestAndSessionURL(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/open/v1/restrict-channels" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			var value map[string]string
			if err := json.Unmarshal(body, &value); err != nil || value["targetChannelId"] != "viewer" {
				t.Fatalf("unexpected restriction body: %s", body)
			}
			return response(`{"code":200,"message":null,"content":null}`, http.StatusOK), nil
		}
		return response(`{"code":200,"message":null,"content":{"url":"wss://socket.example/session"}}`, http.StatusOK), nil
	})
	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token"))
	if err := client.AddRestriction(context.Background(), "viewer"); err != nil {
		t.Fatal(err)
	}
	session, err := client.CreateUserSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if session.URL != "wss://socket.example/session" {
		t.Fatalf("unexpected session URL: %+v", session)
	}
}
