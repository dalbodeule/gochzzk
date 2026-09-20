package chzzk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// User is the authenticated user's CHZZK channel.
type User struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
}

// Channel contains public channel information.
type Channel struct {
	ChannelID       string `json:"channelId"`
	ChannelName     string `json:"channelName"`
	ChannelImageURL string `json:"channelImageUrl"`
	FollowerCount   int    `json:"followerCount"`
	VerifiedMark    bool   `json:"verifiedMark"`
}

// StreamingRole describes a channel manager.
type StreamingRole struct {
	ManagerChannelID   string `json:"managerChannelId"`
	ManagerChannelName string `json:"managerChannelName"`
	UserRole           string `json:"userRole"`
	CreatedDate        string `json:"createdDate"`
}

// Follower is a channel follower.
type Follower struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	CreatedDate string `json:"createdDate"`
}

// Subscriber is a channel subscriber.
type Subscriber struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	Month       int    `json:"month"`
	TierNo      int    `json:"tierNo"`
	CreatedDate string `json:"createdDate"`
}

// PageResult is the paginated content returned by follower/subscriber APIs.
type PageResult[T any] struct {
	Data []T `json:"data"`
}

// Category is a CHZZK content category.
type Category struct {
	CategoryType   string `json:"categoryType"`
	CategoryID     string `json:"categoryId"`
	CategoryValue  string `json:"categoryValue"`
	PosterImageURL string `json:"posterImageUrl"`
}

// Live describes a currently running broadcast.
type Live struct {
	LiveID                int64    `json:"liveId"`
	LiveTitle             string   `json:"liveTitle"`
	LiveThumbnailImageURL string   `json:"liveThumbnailImageUrl"`
	ConcurrentUserCount   int      `json:"concurrentUserCount"`
	OpenDate              string   `json:"openDate"`
	Adult                 bool     `json:"adult"`
	Tags                  []string `json:"tags"`
	CategoryType          string   `json:"categoryType"`
	LiveCategory          string   `json:"liveCategory"`
	LiveCategoryValue     string   `json:"liveCategoryValue"`
	ChannelID             string   `json:"channelId"`
	ChannelName           string   `json:"channelName"`
	ChannelImageURL       string   `json:"channelImageUrl"`
}

// LivesPage is the cursor-paginated live list.
type LivesPage struct {
	Data []Live   `json:"data"`
	Page LivePage `json:"page"`
}
type LivePage struct {
	Next string `json:"next"`
}

// StreamKey contains the stream key for the authenticated streaming channel.
type StreamKey struct {
	StreamKey string `json:"streamKey"`
}

// LiveSetting is the current broadcast setting.
type LiveSetting struct {
	DefaultLiveTitle string   `json:"defaultLiveTitle"`
	Category         Category `json:"category"`
	Tags             []string `json:"tags"`
}

// LiveSettingPatch contains fields to update. Nil fields are omitted.
type LiveSettingPatch struct {
	DefaultLiveTitle *string   `json:"defaultLiveTitle,omitempty"`
	CategoryType     *string   `json:"categoryType,omitempty"`
	CategoryID       *string   `json:"categoryId,omitempty"`
	Tags             *[]string `json:"tags,omitempty"`
}

// SendChatRequest sends a new chat message.
type SendChatRequest struct {
	Message string `json:"message"`
}
type ChatMessage struct {
	MessageID string `json:"messageId"`
}

// NoticeRequest creates a notice from a new message or an existing message ID.
type NoticeRequest struct {
	Message   string `json:"message,omitempty"`
	MessageID string `json:"messageId,omitempty"`
}

// ChatSettings contains the channel's chat policy.
type ChatSettings struct {
	ChatAvailableCondition        string `json:"chatAvailableCondition"`
	ChatAvailableGroup            string `json:"chatAvailableGroup"`
	MinFollowerMinute             int    `json:"minFollowerMinute"`
	AllowSubscriberInFollowerMode bool   `json:"allowSubscriberInFollowerMode"`
	ChatSlowModeSec               *int   `json:"chatSlowModeSec"`
	ChatEmojiMode                 *bool  `json:"chatEmojiMode"`
}

// BlindMessageRequest identifies a message to hide.
type BlindMessageRequest struct {
	ChatChannelID   string `json:"chatChannelId"`
	MessageTime     int64  `json:"messageTime"`
	SenderChannelID string `json:"senderChannelId"`
}

// Restriction identifies a channel that is restricted from activity.
type Restriction struct {
	RestrictedChannelID   string `json:"restrictedChannelId"`
	RestrictedChannelName string `json:"restrictedChannelName"`
	CreatedDate           string `json:"createdDate"`
	ReleaseDate           string `json:"releaseDate"`
}

// RestrictionPage is the cursor-paginated restriction list.
type RestrictionPage struct {
	Data []Restriction     `json:"data"`
	Page RestrictionCursor `json:"page"`
}

// RestrictionCursor contains the cursor for the next restriction page.
type RestrictionCursor struct {
	Next string `json:"next"`
}

// SessionURL is a short-lived Socket.IO connection URL.
type SessionURL struct {
	URL string `json:"url"`
}

// Session describes a previously created Socket.IO session.
type Session struct {
	SessionKey       string            `json:"sessionKey"`
	ConnectedDate    string            `json:"connectedDate"`
	DisconnectedDate string            `json:"disconnectedDate"`
	SubscribedEvents []SubscribedEvent `json:"subscribedEvents"`
}

// SubscribedEvent describes an event subscription attached to a session.
type SubscribedEvent struct {
	EventType string `json:"eventType"`
	ChannelID string `json:"channelId"`
}

// SessionPage is a page of session history.
type SessionPage struct {
	Data []Session `json:"data"`
}

// SessionEvent is the normalized event delivered by a CHZZK Socket.IO session.
type SessionEvent struct {
	EventType string
	Data      json.RawMessage
}

const (
	SessionEventSystem       = "SYSTEM"
	SessionEventChat         = "CHAT"
	SessionEventDonation     = "DONATION"
	SessionEventSubscription = "SUBSCRIPTION"
)

// SystemEvent is delivered for connected/subscribed/unsubscribed/revoked events.
type SystemEvent struct {
	Type string          `json:"type"`
	Data SystemEventData `json:"data"`
}
type SystemEventData struct {
	SessionKey string `json:"sessionKey,omitempty"`
	EventType  string `json:"eventType,omitempty"`
	ChannelID  string `json:"channelId,omitempty"`
}

// ChatEvent is a live chat event.
type ChatEvent struct {
	ChannelID       string            `json:"channelId"`
	SenderChannelID string            `json:"senderChannelId"`
	ChatChannelID   string            `json:"chatChannelId"`
	Profile         ChatProfile       `json:"profile"`
	UserRoleCode    string            `json:"userRoleCode"`
	Content         string            `json:"content"`
	Emojis          map[string]string `json:"emojis"`
	MessageTime     int64             `json:"messageTime"`
	EventSentAt     string            `json:"eventSentAt,omitempty"`
}
type ChatProfile struct {
	Nickname     string      `json:"nickname"`
	Badges       []ChatBadge `json:"badges"`
	VerifiedMark bool        `json:"verifiedMark"`
	UserRoleCode string      `json:"userRoleCode,omitempty"`
}

// EffectiveUserRoleCode returns the observed nested role first, then the
// top-level role documented by the official API.
func (e ChatEvent) EffectiveUserRoleCode() string {
	if e.Profile.UserRoleCode != "" {
		return e.Profile.UserRoleCode
	}
	return e.UserRoleCode
}

// MessageTimeUTC converts the event's Unix epoch milliseconds to UTC.
// Unlike EventSentAt, MessageTime is an absolute timestamp and has no KST-only interpretation.
func (e ChatEvent) MessageTimeUTC() time.Time {
	return time.UnixMilli(e.MessageTime).UTC()
}

type ChatBadge struct {
	BadgeNo     any    `json:"badgeNo,omitempty"`
	BadgeID     string `json:"badgeId,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Activated   *bool  `json:"activated,omitempty"`
}

// DonationEvent is a live donation event.
type DonationEvent struct {
	DonationType     string            `json:"donationType"`
	ChannelID        string            `json:"channelId"`
	DonatorChannelID string            `json:"donatorChannelId"`
	DonatorNickname  string            `json:"donatorNickname"`
	PayAmount        DonationAmount    `json:"payAmount"`
	DonationText     string            `json:"donationText"`
	Emojis           map[string]string `json:"emojis"`
	EventSentAt      string            `json:"eventSentAt,omitempty"`
}

// DonationAmount accepts both the documented JSON string and observed number.
type DonationAmount int64

func (a *DonationAmount) UnmarshalJSON(data []byte) error {
	raw := string(bytes.TrimSpace(data))
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		raw = value
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("decode donation amount %q: %w", raw, err)
	}
	*a = DonationAmount(value)
	return nil
}

func (a DonationAmount) String() string { return strconv.FormatInt(int64(a), 10) }

// SubscriptionEvent is a new subscription event.
type SubscriptionEvent struct {
	ChannelID           string `json:"channelId"`
	SubscriberChannelID string `json:"subscriberChannelId"`
	SubscriberNickname  string `json:"subscriberNickname"`
	TierNo              int    `json:"tierNo"`
	TierName            string `json:"tierName"`
	Month               int    `json:"month"`
	EventSentAt         string `json:"eventSentAt,omitempty"`
}

// ParseEventSentAt parses the observed offset-less KST event timestamp.
func ParseEventSentAt(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("chzzk: eventSentAt is empty")
	}
	return time.ParseInLocation("2006-01-02T15:04:05.999999999", value, time.FixedZone("KST", 9*60*60))
}

// ChatSettingsPatch contains fields to update. Nil fields are omitted.
type ChatSettingsPatch struct {
	ChatAvailableCondition        *string `json:"chatAvailableCondition,omitempty"`
	ChatAvailableGroup            *string `json:"chatAvailableGroup,omitempty"`
	MinFollowerMinute             *int    `json:"minFollowerMinute,omitempty"`
	AllowSubscriberInFollowerMode *bool   `json:"allowSubscriberInFollowerMode,omitempty"`
	ChatSlowModeSec               *int    `json:"chatSlowModeSec,omitempty"`
	ChatEmojiMode                 *bool   `json:"chatEmojiMode,omitempty"`
}
