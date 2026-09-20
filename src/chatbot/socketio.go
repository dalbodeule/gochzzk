package chatbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	api "github.com/dalbodeule/gochzzk/src/api"
	"github.com/gorilla/websocket"
)

// EventHandler handles normalized events from a CHZZK Socket.IO session.
type EventHandler func(api.SessionEvent) error

// SocketSession is a small Socket.IO v2-compatible transport for CHZZK sessions.
// The API returns a short-lived URL; pass that URL to NewSocketSession and call Run.
type SocketSession struct {
	sessionURL string
	dialer     *websocket.Dialer
}

const socketOpenTimeout = 30 * time.Second

type engineOpenPacket struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int64    `json:"pingInterval"`
	PingTimeout  int64    `json:"pingTimeout"`
}

// SocketOption configures a SocketSession before it is run.
type SocketOption func(*SocketSession)

// WithDialer replaces the WebSocket dialer used by a SocketSession.
func WithDialer(dialer *websocket.Dialer) SocketOption {
	return func(s *SocketSession) { s.dialer = dialer }
}

// NewSocketSession creates a Socket.IO session transport from a URL returned by the API.
func NewSocketSession(sessionURL string, options ...SocketOption) *SocketSession {
	s := &SocketSession{sessionURL: sessionURL, dialer: websocket.DefaultDialer}
	for _, option := range options {
		option(s)
	}
	return s
}

// Run connects using the Engine.IO v3 transport used by the documented Socket.IO 2.x client,
// then invokes handler for SYSTEM, CHAT, DONATION, and SUBSCRIPTION events until disconnect.
func (s *SocketSession) Run(ctx context.Context, handler EventHandler) error {
	socketURL, err := buildSocketURL(s.sessionURL)
	if err != nil {
		return err
	}
	if s.dialer == nil {
		return fmt.Errorf("chzzk: websocket dialer is nil")
	}
	conn, _, err := s.dialer.DialContext(ctx, socketURL, nil)
	if err != nil {
		return fmt.Errorf("dial chzzk socket: %w", err)
	}
	defer conn.Close()
	cancelWatchDone := make(chan struct{})
	defer close(cancelWatchDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-cancelWatchDone:
		}
	}()
	conn.SetReadLimit(1 << 20)
	openDeadline := time.Now().Add(socketOpenTimeout)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(openDeadline) {
		openDeadline = deadline
	}
	if err := conn.SetReadDeadline(openDeadline); err != nil {
		return fmt.Errorf("set chzzk socket handshake deadline: %w", err)
	}
	_, opening, err := conn.ReadMessage()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("read chzzk socket handshake: %w", err)
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("clear chzzk socket handshake deadline: %w", err)
	}
	if len(opening) == 0 || opening[0] != '0' {
		return fmt.Errorf("unexpected chzzk socket handshake packet: %q", opening)
	}
	var handshake engineOpenPacket
	if err := json.Unmarshal(opening[1:], &handshake); err != nil {
		return fmt.Errorf("decode chzzk socket handshake: %w", err)
	}
	pingInterval := time.Duration(handshake.PingInterval) * time.Millisecond
	pingTimeout := time.Duration(handshake.PingTimeout) * time.Millisecond
	if pingInterval <= 0 {
		pingInterval = 25 * time.Second
	}
	if pingTimeout <= 0 {
		pingTimeout = 60 * time.Second
	}

	stop := make(chan struct{})
	defer close(stop)
	pong := make(chan struct{}, 1)
	heartbeatErr := make(chan error, 1)
	go func() {
		if err := runHeartbeat(ctx, conn, pingInterval, pingTimeout, pong, stop); err != nil {
			select {
			case heartbeatErr <- err:
			default:
			}
		}
		_ = conn.Close()
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			select {
			case heartbeat := <-heartbeatErr:
				return heartbeat
			default:
			}
			return fmt.Errorf("read chzzk socket: %w", err)
		}
		packet := string(message)
		switch {
		case strings.HasPrefix(packet, "0"):
			continue // Engine.IO open packet.
		case packet == "3":
			select {
			case pong <- struct{}{}:
			default:
			}
		case strings.HasPrefix(packet, "42"):
			event, err := decodeSocketIOEvent([]byte(packet[2:]))
			if err != nil {
				return err
			}
			if handler != nil {
				if err := handler(event); err != nil {
					return err
				}
			}
		}
	}
}

func buildSocketURL(sessionURL string) (string, error) {
	u, err := url.Parse(sessionURL)
	if err != nil || u.Host == "" {
		return "", errors.New("chzzk: invalid session URL")
	}
	if u.User != nil || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("chzzk: invalid session URL components")
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "wss":
	default:
		return "", errors.New("chzzk: session URL must use HTTPS or WSS")
	}
	if err := validateSessionEndpoint(u); err != nil {
		return "", err
	}
	q := u.Query()
	authValues, ok := q["auth"]
	if !ok || len(authValues) != 1 || authValues[0] == "" {
		return "", errors.New("chzzk: session URL must contain exactly one auth token")
	}
	u.Path = "/socket.io/"
	u.RawPath = ""
	q.Set("EIO", "3")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func validateSessionEndpoint(u *url.URL) error {
	if port := u.Port(); port != "" && port != "443" {
		return errors.New("chzzk: session URL must use port 443")
	}
	host := strings.ToLower(u.Hostname())
	const prefix = "ssio"
	const suffix = ".nchat.naver.com"
	if !strings.HasPrefix(host, prefix) || !strings.HasSuffix(host, suffix) {
		return errors.New("chzzk: session URL must use a CHZZK session host")
	}
	numberText := strings.TrimSuffix(strings.TrimPrefix(host, prefix), suffix)
	if len(numberText) != 2 {
		return errors.New("chzzk: invalid CHZZK session host number")
	}
	number, err := strconv.Atoi(numberText)
	if err != nil || number < 1 || number > 29 {
		return errors.New("chzzk: invalid CHZZK session host number")
	}
	return nil
}

func runHeartbeat(ctx context.Context, conn *websocket.Conn, interval, timeout time.Duration, pong <-chan struct{}, stop <-chan struct{}) error {
	for {
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-stop:
			timer.Stop()
			return nil
		case <-timer.C:
		}
		// A pong only acknowledges the ping sent in this cycle. Discard any
		// duplicate or late pong retained from the previous cycle.
		select {
		case <-pong:
		default:
		}
		if err := writePacket(conn, "2"); err != nil {
			return fmt.Errorf("send chzzk socket ping: %w", err)
		}
		timer.Reset(timeout)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-stop:
			timer.Stop()
			return nil
		case <-pong:
			timer.Stop()
		case <-timer.C:
			return errors.New("chzzk socket heartbeat timed out")
		}
	}
}

func writePacket(conn *websocket.Conn, packet string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(packet))
}

func decodeSocketIOEvent(payload []byte) (api.SessionEvent, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(payload, &parts); err != nil {
		return api.SessionEvent{}, fmt.Errorf("decode socket event: %w", err)
	}
	if len(parts) < 2 {
		return api.SessionEvent{}, fmt.Errorf("decode socket event: expected event name and data")
	}
	var eventType string
	if err := json.Unmarshal(parts[0], &eventType); err != nil {
		return api.SessionEvent{}, fmt.Errorf("decode socket event name: %w", err)
	}
	data := parts[1]
	if len(data) > 0 && data[0] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return api.SessionEvent{}, fmt.Errorf("decode socket event string: %w", err)
		}
		if !json.Valid([]byte(encoded)) {
			return api.SessionEvent{}, fmt.Errorf("decode socket event string: invalid JSON")
		}
		data = json.RawMessage(encoded)
	}
	return api.SessionEvent{EventType: eventType, Data: data}, nil
}

// DecodeEvent decodes a normalized event payload into the corresponding typed value.
func DecodeEvent[T any](event api.SessionEvent) (T, error) {
	var value T
	err := json.Unmarshal(event.Data, &value)
	return value, err
}

// ChatBot provides typed callbacks over a SocketSession, making a simple chatbot easy to compose.
type ChatBot struct {
	Session        SessionRunner
	OnSystem       func(api.SystemEvent) error
	OnChat         func(api.ChatEvent) error
	OnDonation     func(api.DonationEvent) error
	OnSubscription func(api.SubscriptionEvent) error
}

// Run starts the chatbot event loop.
func (b *ChatBot) Run(ctx context.Context) error {
	return b.runWithRunner(ctx)
}
