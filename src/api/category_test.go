package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchCategories(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("query") != "게임" || r.URL.Query().Get("size") != "10" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		return response(`{"code":200,"message":null,"content":[{"categoryId":"1","categoryValue":"게임"}]}`, http.StatusOK), nil
	})
	categories, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithClientCredentials("id", "secret")).SearchCategories(context.Background(), "게임", 10)
	if err != nil || len(categories) != 1 || categories[0].CategoryValue != "게임" {
		t.Fatalf("unexpected categories: %+v, err=%v", categories, err)
	}
}
