package chzzk

import (
	"context"
	"net/http"
)

// SendChat sends a chat message and returns its message ID.
func (c *Client) SendChat(ctx context.Context, request SendChatRequest) (ChatMessage, error) {
	if err := ValidateChatMessage(request.Message); err != nil {
		return ChatMessage{}, err
	}
	var v ChatMessage
	err := c.do(ctx, http.MethodPost, "/open/v1/chats/send", nil, request, authAccessToken, &v)
	return v, err
}

// RegisterNotice registers a new or existing message as the channel notice.
func (c *Client) RegisterNotice(ctx context.Context, request NoticeRequest) error {
	if request.Message != "" {
		if err := ValidateChatMessage(request.Message); err != nil {
			return err
		}
	}
	return c.do(ctx, http.MethodPost, "/open/v1/chats/notice", nil, request, authAccessToken, nil)
}

// GetChatSettings returns the authenticated channel's chat settings.
func (c *Client) GetChatSettings(ctx context.Context) (ChatSettings, error) {
	var v ChatSettings
	err := c.do(ctx, http.MethodGet, "/open/v1/chats/settings", nil, nil, authAccessToken, &v)
	return v, err
}

// UpdateChatSettings updates the authenticated channel's chat settings.
func (c *Client) UpdateChatSettings(ctx context.Context, patch ChatSettingsPatch) error {
	return c.do(ctx, http.MethodPut, "/open/v1/chats/settings", nil, patch, authAccessToken, nil)
}

// BlindMessage hides a chat message.
func (c *Client) BlindMessage(ctx context.Context, request BlindMessageRequest) error {
	return c.do(ctx, http.MethodPost, "/open/v1/chats/blind-message", nil, request, authAccessToken, nil)
}
