package chatbot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	api "github.com/dalbodeule/gochzzk/src/api"
)

// StoredToken is one user's token pair and optional proactive expiry time.
type StoredToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewStoredToken converts an API token response into a token suitable for TokenStore.
func NewStoredToken(token api.Token, now time.Time) StoredToken {
	stored := StoredToken{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}
	seconds, err := strconv.ParseInt(token.ExpiresIn, 10, 64)
	if err == nil && seconds > 0 {
		stored.ExpiresAt = now.Add(time.Duration(seconds) * time.Second)
	}
	return stored
}

// TokenStore persists token pairs independently for each authorized CHZZK user.
// Save must atomically replace both tokens because refresh tokens are single-use.
type TokenStore interface {
	Load(context.Context, string) (StoredToken, error)
	Save(context.Context, string, StoredToken) error
}

var ErrTokenNotFound = errors.New("chzzk chatbot: token not found")
var ErrBotAlreadyRunning = errors.New("chzzk chatbot: this user bot is already running")

// MemoryTokenStore is a race-safe in-memory TokenStore for tests and single-process bots.
// Production services should use an encrypted durable implementation.
type MemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]StoredToken
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{tokens: make(map[string]StoredToken)}
}

func (s *MemoryTokenStore) Load(_ context.Context, userID string) (StoredToken, error) {
	s.mu.RLock()
	token, ok := s.tokens[userID]
	s.mu.RUnlock()
	if !ok {
		return StoredToken{}, ErrTokenNotFound
	}
	return token, nil
}

func (s *MemoryTokenStore) Save(_ context.Context, userID string, token StoredToken) error {
	s.mu.Lock()
	if s.tokens == nil {
		s.tokens = make(map[string]StoredToken)
	}
	s.tokens[userID] = token
	s.mu.Unlock()
	return nil
}

type SessionRunner interface {
	Run(context.Context, EventHandler) error
}
type SocketFactory func(string) SessionRunner

// UserChatBot runs one user-authorized CHZZK session and subscribes that user to
// chat, donation, and subscription events. Create one UserChatBot per authorized user.
type UserChatBot struct {
	runMu sync.Mutex

	UserID        string
	ClientID      string
	ClientSecret  string
	Tokens        TokenStore
	APIOptions    []api.Option
	SocketFactory SocketFactory
	RefreshBefore time.Duration
	Now           func() time.Time

	OnSystem       func(api.SystemEvent) error
	OnChat         func(api.ChatEvent) error
	OnDonation     func(api.DonationEvent) error
	OnSubscription func(api.SubscriptionEvent) error
}

// Run loads this user's token, refreshes it when necessary, creates a user session,
// and subscribes CHAT, DONATION, and SUBSCRIPTION after SYSTEM/connected.
func (b *UserChatBot) Run(ctx context.Context) error {
	if err := b.validate(); err != nil {
		return err
	}
	if !b.runMu.TryLock() {
		return ErrBotAlreadyRunning
	}
	defer b.runMu.Unlock()
	token, err := b.Tokens.Load(ctx, b.UserID)
	if err != nil {
		return err
	}
	if b.tokenExpiresSoon(token) {
		token, err = b.refresh(ctx, token)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		err = b.runSession(ctx, token)
		if err == nil || ctx.Err() != nil || !isInvalidToken(err) || attempt == 1 {
			return err
		}
		token, err = b.refresh(ctx, token)
		if err != nil {
			return err
		}
	}
	return err
}

func (b *UserChatBot) validate() error {
	if b == nil {
		return errors.New("chzzk chatbot: nil UserChatBot")
	}
	if b.UserID == "" || b.ClientID == "" || b.ClientSecret == "" {
		return errors.New("chzzk chatbot: user ID and client credentials are required")
	}
	if b.Tokens == nil {
		return errors.New("chzzk chatbot: token store is required")
	}
	return nil
}

func (b *UserChatBot) tokenExpiresSoon(token StoredToken) bool {
	if token.ExpiresAt.IsZero() {
		return false
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	skew := b.RefreshBefore
	if skew <= 0 {
		skew = time.Minute
	}
	return !token.ExpiresAt.After(now().Add(skew))
}

func (b *UserChatBot) refresh(ctx context.Context, old StoredToken) (StoredToken, error) {
	if old.RefreshToken == "" {
		return StoredToken{}, errors.New("chzzk chatbot: refresh token is required")
	}
	client := api.New(b.apiOptions("")...)
	response, err := client.RefreshToken(ctx, old.RefreshToken)
	if err != nil {
		return StoredToken{}, fmt.Errorf("refresh user token: %w", err)
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	updated := NewStoredToken(response, now())
	if err := b.Tokens.Save(ctx, b.UserID, updated); err != nil {
		return StoredToken{}, fmt.Errorf("save refreshed user token: %w", err)
	}
	return updated, nil
}

func (b *UserChatBot) runSession(ctx context.Context, token StoredToken) error {
	if token.AccessToken == "" {
		return errors.New("chzzk chatbot: access token is required")
	}
	client := api.New(b.apiOptions(token.AccessToken)...)
	session, err := client.CreateUserSession(ctx)
	if err != nil {
		return err
	}
	factory := b.SocketFactory
	if factory == nil {
		factory = func(rawURL string) SessionRunner { return NewSocketSession(rawURL) }
	}
	runner := factory(session.URL)
	if runner == nil {
		return errors.New("chzzk chatbot: socket factory returned nil")
	}
	bot := ChatBot{
		Session: runner,
		OnSystem: func(event api.SystemEvent) error {
			if event.Type == "connected" {
				if err := client.SubscribeChat(ctx, event.Data.SessionKey); err != nil {
					return err
				}
				if err := client.SubscribeDonation(ctx, event.Data.SessionKey); err != nil {
					return err
				}
				if err := client.SubscribeSubscription(ctx, event.Data.SessionKey); err != nil {
					return err
				}
			}
			if b.OnSystem != nil {
				return b.OnSystem(event)
			}
			return nil
		},
		OnChat: b.OnChat, OnDonation: b.OnDonation, OnSubscription: b.OnSubscription,
	}
	return bot.runWithRunner(ctx)
}

func (b *UserChatBot) apiOptions(accessToken string) []api.Option {
	options := append([]api.Option(nil), b.APIOptions...)
	options = append(options, api.WithClientCredentials(b.ClientID, b.ClientSecret))
	if accessToken != "" {
		options = append(options, api.WithAccessToken(accessToken))
	}
	return options
}

func isInvalidToken(err error) bool {
	var apiErr *api.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 401 && (apiErr.Message == "INVALID_TOKEN" || apiErr.Code == 401)
}

func (b *ChatBot) runWithRunner(ctx context.Context) error {
	if b == nil || b.Session == nil {
		return errors.New("chzzk: chatbot session is nil")
	}
	return b.Session.Run(ctx, b.handle)
}

func (b *ChatBot) handle(event api.SessionEvent) error {
	switch event.EventType {
	case api.SessionEventSystem:
		if b.OnSystem == nil {
			return nil
		}
		value, err := DecodeEvent[api.SystemEvent](event)
		if err != nil {
			return err
		}
		return b.OnSystem(value)
	case api.SessionEventChat:
		if b.OnChat == nil {
			return nil
		}
		value, err := DecodeEvent[api.ChatEvent](event)
		if err != nil {
			return err
		}
		return b.OnChat(value)
	case api.SessionEventDonation:
		if b.OnDonation == nil {
			return nil
		}
		value, err := DecodeEvent[api.DonationEvent](event)
		if err != nil {
			return err
		}
		return b.OnDonation(value)
	case api.SessionEventSubscription:
		if b.OnSubscription == nil {
			return nil
		}
		value, err := DecodeEvent[api.SubscriptionEvent](event)
		if err != nil {
			return err
		}
		return b.OnSubscription(value)
	default:
		return nil
	}
}
