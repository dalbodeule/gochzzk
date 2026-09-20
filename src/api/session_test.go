package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestSubscribeDonation(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/open/v1/sessions/events/subscribe/donation" || r.URL.Query().Get("sessionKey") != "session" {
			t.Fatalf("unexpected session request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		return response(`{"code":200,"message":null,"content":null}`, http.StatusOK), nil
	})
	err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).SubscribeDonation(context.Background(), "session")
	if err != nil {
		t.Fatal(err)
	}
}
