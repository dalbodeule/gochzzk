package unofficial

import (
	"context"
	"fmt"
	"net/url"
)

type LiveStatus struct {
	LiveTitle           string  `json:"liveTitle"`
	Status              string  `json:"status"`
	ConcurrentUserCount int     `json:"concurrentUserCount"`
	AccumulateCount     int     `json:"accumulateCount"`
	PaidPromotion       bool    `json:"paidPromotion"`
	Adult               bool    `json:"adult"`
	KROnlyViewing       bool    `json:"krOnlyViewing"`
	OpenDate            string  `json:"openDate"`
	CloseDate           *string `json:"closeDate"`
	ClipActive          bool    `json:"clipActive"`
	ChatChannelID       string  `json:"chatChannelId"`
}

// GetLiveStatus calls CHZZK's undocumented polling endpoint for one channel.
func (c *Client) GetLiveStatus(ctx context.Context, channelID string) (*LiveStatus, error) {
	id, err := escaped(channelID)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(c.liveBaseURL + "/polling/v3/channels/" + id + "/live-status")
	if err != nil {
		return nil, fmt.Errorf("parse unofficial live URL: %w", err)
	}
	q := u.Query()
	q.Set("includePlayerRecommendContent", "false")
	u.RawQuery = q.Encode()
	var status LiveStatus
	found, err := c.get(ctx, u.String(), &status)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &status, nil
}

// GetChatChannelID returns the current live's chat channel ID.
func (c *Client) GetChatChannelID(ctx context.Context, channelID string) (string, error) {
	status, err := c.GetLiveStatus(ctx, channelID)
	if err != nil || status == nil {
		return "", err
	}
	return status.ChatChannelID, nil
}
