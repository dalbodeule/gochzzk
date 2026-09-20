package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestGetChannelsEncodesRepeatedIDs(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		ids := r.URL.Query()["channelIds"]
		if len(ids) != 2 || ids[0] != "one" || ids[1] != "two two" {
			t.Fatalf("channelIds = %#v", ids)
		}
		return response(`{"code":200,"message":null,"content":[{"channelId":"one","channelName":"One"}]}`, http.StatusOK), nil
	})
	channels, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithClientCredentials("id", "secret")).GetChannels(context.Background(), []string{"one", "two two"})
	if err != nil || len(channels) != 1 || channels[0].ChannelName != "One" {
		t.Fatalf("unexpected channels: %+v, err=%v", channels, err)
	}
}
