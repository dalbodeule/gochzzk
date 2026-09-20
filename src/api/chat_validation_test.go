package chzzk

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestValidateChatMessageUsesUnicodeRunes(t *testing.T) {
	if err := ValidateChatMessage(strings.Repeat("가", 100)); err != nil {
		t.Fatal(err)
	}
	err := ValidateChatMessage(strings.Repeat("가", 101))
	var lengthErr *ChatMessageLengthError
	if !errors.As(err, &lengthErr) || lengthErr.Length != 101 || lengthErr.Limit != 100 {
		t.Fatalf("unexpected length error: %v", err)
	}
}

func TestTruncateChatMessageDoesNotSplitUTF8(t *testing.T) {
	got := TruncateChatMessage(strings.Repeat("가", 100) + "abc")
	if utf8.RuneCountInString(got) != 100 || !utf8.ValidString(got) {
		t.Fatalf("unexpected truncated message: %q", got)
	}
}

func TestSendChatRejectsBeforeHTTPCall(t *testing.T) {
	called := false
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return response(`{}`, http.StatusOK), nil
	})
	_, err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).SendChat(context.Background(), SendChatRequest{Message: strings.Repeat("a", 101)})
	if err == nil || called {
		t.Fatalf("expected local validation error, called=%v err=%v", called, err)
	}
}

func TestRegisterNoticeRejectsLongMessageBeforeHTTPCall(t *testing.T) {
	called := false
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return response(`{}`, http.StatusOK), nil
	})
	err := New(WithHTTPClient(&http.Client{Transport: transport}), WithAccessToken("token")).RegisterNotice(context.Background(), NoticeRequest{Message: strings.Repeat("a", 101)})
	if err == nil || called {
		t.Fatalf("expected local validation error, called=%v err=%v", called, err)
	}
}

func TestValidateChatMessageRejectsInvalidUTF8(t *testing.T) {
	if err := ValidateChatMessage(string([]byte{0xff})); !errors.Is(err, ErrInvalidChatMessageUTF8) {
		t.Fatalf("unexpected UTF-8 error: %v", err)
	}
}
