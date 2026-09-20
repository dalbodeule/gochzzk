package chatbot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	api "github.com/dalbodeule/gochzzk/src/api"
)

type botRoundTripper func(*http.Request) (*http.Response, error)

func (f botRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func botResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

type connectedRunner struct{}

func (connectedRunner) Run(_ context.Context, handler EventHandler) error {
	return handler(api.SessionEvent{EventType: api.SessionEventSystem, Data: []byte(`{"type":"connected","data":{"sessionKey":"session-key"}}`)})
}

func TestUserChatBotRefreshesAndSubscribesAllEvents(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	store := NewMemoryTokenStore()
	if err := store.Save(context.Background(), "user", StoredToken{AccessToken: "old", RefreshToken: "old-refresh", ExpiresAt: now.Add(-time.Minute)}); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	calls := make(map[string]int)
	transport := botRoundTripper(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		calls[r.URL.Path]++
		mu.Unlock()
		switch r.URL.Path {
		case "/auth/v1/token":
			return botResponse(`{"code":200,"message":null,"content":{"accessToken":"new","refreshToken":"new-refresh","tokenType":"Bearer","expiresIn":"86400"}}`), nil
		case "/open/v1/sessions/auth":
			if r.Header.Get("Authorization") != "Bearer new" {
				t.Fatalf("session used wrong token: %q", r.Header.Get("Authorization"))
			}
			return botResponse(`{"code":200,"message":null,"content":{"url":"wss://socket.example"}}`), nil
		default:
			if !strings.HasPrefix(r.URL.Path, "/open/v1/sessions/events/subscribe/") {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			return botResponse(`{"code":200,"message":null,"content":null}`), nil
		}
	})

	bot := UserChatBot{
		UserID: "user", ClientID: "client", ClientSecret: "secret", Tokens: store,
		APIOptions:    []api.Option{api.WithHTTPClient(&http.Client{Transport: transport}), api.WithBaseURL("https://test.invalid")},
		SocketFactory: func(string) SessionRunner { return connectedRunner{} }, Now: func() time.Time { return now },
	}
	if err := bot.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/open/v1/sessions/events/subscribe/chat", "/open/v1/sessions/events/subscribe/donation", "/open/v1/sessions/events/subscribe/subscription"} {
		if calls[path] != 1 {
			t.Fatalf("%s calls = %d", path, calls[path])
		}
	}
	stored, err := store.Load(context.Background(), "user")
	if err != nil || stored.AccessToken != "new" || stored.RefreshToken != "new-refresh" {
		t.Fatalf("unexpected stored token: %+v, err=%v", stored, err)
	}
}

func TestMemoryTokenStoreConcurrentAccess(t *testing.T) {
	store := NewMemoryTokenStore()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Save(context.Background(), "user", StoredToken{AccessToken: "token"})
			_, _ = store.Load(context.Background(), "user")
		}()
	}
	wg.Wait()
}
