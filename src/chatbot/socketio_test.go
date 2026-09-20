package chatbot

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	api "github.com/dalbodeule/gochzzk/src/api"
	"github.com/gorilla/websocket"
)

const localSessionURL = "https://ssio08.nchat.naver.com:443?auth=test-token"

func newSocketTestServer(t *testing.T, handler func(*websocket.Conn) error) (*httptest.Server, *websocket.Dialer, <-chan error) {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local TCP listener is unavailable: %v", err)
	}
	serverResult := make(chan error, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverResult <- fmt.Errorf("upgrade test websocket: %w", err)
			return
		}
		defer conn.Close()
		serverResult <- handler(conn)
	}))
	server.Listener = listener
	server.StartTLS()

	target := server.Listener.Addr().String()
	dialer := *websocket.DefaultDialer
	dialer.Proxy = nil
	dialer.NetDialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, target)
	}
	// The URL remains the production CHZZK hostname while the test-only dialer
	// routes the connection to httptest's self-signed local TLS server.
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // test server only
	return server, &dialer, serverResult
}

func assertServerResult(t *testing.T, serverResult <-chan error) {
	t.Helper()
	select {
	case err := <-serverResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("test websocket server did not finish")
	}
}

func TestSocketSessionRunUsesEngineIOV3WithoutRootConnect(t *testing.T) {
	secondPongSent := make(chan struct{}, 1)
	releaseServer := make(chan struct{})
	server, dialer, serverResult := newSocketTestServer(t, func(conn *websocket.Conn) error {
		opening := `0{"sid":"session","upgrades":[],"pingInterval":15,"pingTimeout":200}`
		if err := conn.WriteMessage(websocket.TextMessage, []byte(opening)); err != nil {
			return err
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(`42["CHAT","{\"channelId\":\"channel\",\"content\":\"hello\"}"]`)); err != nil {
			return err
		}
		for pingCount := 0; pingCount < 2; pingCount++ {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("read client packet: %w", err)
			}
			if packet := string(message); packet != "2" {
				return fmt.Errorf("unexpected client packet %q; root namespace CONNECT must not be sent", packet)
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte("3")); err != nil {
				return err
			}
		}
		secondPongSent <- struct{}{}
		<-releaseServer
		return nil
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan api.SessionEvent, 1)
	runDone := make(chan error, 1)
	go func() {
		runDone <- NewSocketSession(localSessionURL, WithDialer(dialer)).Run(ctx, func(event api.SessionEvent) error {
			events <- event
			return nil
		})
	}()

	select {
	case event := <-events:
		if event.EventType != api.SessionEventChat || string(event.Data) != `{"channelId":"channel","content":"hello"}` {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive CHAT event")
	}
	select {
	case <-secondPongSent:
	case <-time.After(time.Second):
		t.Fatal("did not receive two client pings")
	}

	cancel()
	select {
	case err := <-runDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}
	close(releaseServer)
	assertServerResult(t, serverResult)
}

func TestSocketSessionRunCancelsWhileWaitingForOpenPacket(t *testing.T) {
	accepted := make(chan struct{}, 1)
	releaseServer := make(chan struct{})
	server, dialer, serverResult := newSocketTestServer(t, func(_ *websocket.Conn) error {
		accepted <- struct{}{}
		<-releaseServer
		return nil
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- NewSocketSession(localSessionURL, WithDialer(dialer)).Run(ctx, nil)
	}()
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("server did not accept connection")
	}
	cancel()
	select {
	case err := <-runDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop while waiting for open packet")
	}
	close(releaseServer)
	assertServerResult(t, serverResult)
}

func TestSocketSessionRunTimesOutWithoutPong(t *testing.T) {
	pingReceived := make(chan struct{}, 1)
	releaseServer := make(chan struct{})
	server, dialer, serverResult := newSocketTestServer(t, func(conn *websocket.Conn) error {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(`0{"sid":"session","upgrades":["websocket"],"pingInterval":10,"pingTimeout":75}`)); err != nil {
			return err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read client ping: %w", err)
		}
		if packet := string(message); packet != "2" {
			return fmt.Errorf("unexpected client packet %q", packet)
		}
		pingReceived <- struct{}{}
		<-releaseServer
		return nil
	})
	defer server.Close()

	runDone := make(chan error, 1)
	go func() {
		runDone <- NewSocketSession(localSessionURL, WithDialer(dialer)).Run(context.Background(), nil)
	}()
	select {
	case <-pingReceived:
	case <-time.After(time.Second):
		t.Fatal("server did not receive heartbeat ping")
	}
	select {
	case err := <-runDone:
		if err == nil || !strings.Contains(err.Error(), "heartbeat timed out") {
			t.Fatalf("Run error = %v, want heartbeat timeout", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not time out without pong")
	}
	close(releaseServer)
	assertServerResult(t, serverResult)
}

func TestEngineOpenPacketHeartbeatValues(t *testing.T) {
	var packet engineOpenPacket
	err := json.Unmarshal([]byte(`{"sid":"session","upgrades":["websocket"],"pingInterval":25000,"pingTimeout":60000}`), &packet)
	if err != nil {
		t.Fatal(err)
	}
	if packet.PingInterval != 25_000 || packet.PingTimeout != 60_000 {
		t.Fatalf("unexpected heartbeat values: %+v", packet)
	}
}

func TestBuildSocketURL(t *testing.T) {
	got, err := buildSocketURL("https://ssio08.nchat.naver.com:443?auth=secret")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(got)
	if err != nil || u.Scheme != "wss" || u.Path != "/socket.io/" || u.Query().Get("auth") != "secret" || u.Query().Get("EIO") != "3" || u.Query().Get("transport") != "websocket" {
		t.Fatalf("unexpected socket URL: %s, err=%v", got, err)
	}
}

func TestBuildSocketURLRejectsUntrustedEndpoints(t *testing.T) {
	tests := []string{
		"http://ssio08.nchat.naver.com:443?auth=secret",
		"https://example.com:443?auth=secret",
		"https://ssio00.nchat.naver.com:443?auth=secret",
		"https://ssio30.nchat.naver.com:443?auth=secret",
		"https://ssio08.nchat.naver.com:80?auth=secret",
		"https://ssio08.nchat.naver.com:443",
		"https://ssio08.nchat.naver.com:443/custom?auth=secret",
	}
	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := buildSocketURL(rawURL); err == nil {
				t.Fatalf("expected rejection for %s", rawURL)
			}
		})
	}
}

func TestDecodeDonationEvent(t *testing.T) {
	event, err := decodeSocketIOEvent([]byte(`["DONATION",{"payAmount":"1000","donationText":"nice"}]`))
	if err != nil {
		t.Fatal(err)
	}
	donation, err := DecodeEvent[api.DonationEvent](event)
	if err != nil || donation.PayAmount.String() != "1000" || donation.DonationText != "nice" {
		t.Fatalf("unexpected donation: %+v, err=%v", donation, err)
	}
}

func TestDecodeDoubleEncodedSocketIOEvent(t *testing.T) {
	event, err := decodeSocketIOEvent([]byte(`["SYSTEM","{\"type\":\"connected\",\"data\":{\"sessionKey\":\"key\"}}"]`))
	if err != nil {
		t.Fatal(err)
	}
	system, err := DecodeEvent[api.SystemEvent](event)
	if err != nil || system.Type != "connected" || system.Data.SessionKey != "key" {
		t.Fatalf("unexpected system event: %+v, err=%v", system, err)
	}
}

func TestDecodeChatEvent(t *testing.T) {
	event, err := decodeSocketIOEvent([]byte(`["CHAT",{"channelId":"channel","content":"hello"}]`))
	if err != nil {
		t.Fatal(err)
	}
	chat, err := DecodeEvent[api.ChatEvent](event)
	if err != nil || chat.ChannelID != "channel" || chat.Content != "hello" {
		t.Fatalf("unexpected chat: %+v, err=%v", chat, err)
	}
}
