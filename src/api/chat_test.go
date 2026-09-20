package chzzk

import (
	"context"
	"net/http"
	"testing"
)

func TestSendChat(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/open/v1/chats/send" {
			t.Fatalf("unexpected chat request")
		}
		return response(`{"code":200,"message":null,"content":{"messageId":"message-1"}}`, http.StatusOK), nil
	})
	message, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).SendChat(context.Background(), SendChatRequest{Message: "hello"})
	if err != nil || message.MessageID != "message-1" {
		t.Fatalf("unexpected message: %+v, err=%v", message, err)
	}
}
