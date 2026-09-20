package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestGetRestrictions(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("size") != "30" || r.URL.Query().Get("next") != "cursor" {
			t.Fatalf("unexpected restriction query: %s", r.URL.RawQuery)
		}
		return response(`{"code":200,"message":null,"content":{"data":[{"restrictedChannelId":"viewer"}],"page":{"next":"next-cursor"}}}`, http.StatusOK), nil
	})
	page, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).GetRestrictions(context.Background(), 30, "cursor")
	if err != nil || len(page.Data) != 1 || page.Page.Next != "next-cursor" {
		t.Fatalf("unexpected restrictions: %+v, err=%v", page, err)
	}
}
