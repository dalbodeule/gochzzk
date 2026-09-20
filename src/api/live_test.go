package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestGetLiveSetting(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/open/v1/lives/setting" || r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected live setting request")
		}
		return response(`{"code":200,"message":null,"content":{"defaultLiveTitle":"오늘 방송","tags":["tag"]}}`, http.StatusOK), nil
	})
	setting, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).GetLiveSetting(context.Background())
	if err != nil || setting.DefaultLiveTitle != "오늘 방송" || len(setting.Tags) != 1 {
		t.Fatalf("unexpected setting: %+v, err=%v", setting, err)
	}
}
