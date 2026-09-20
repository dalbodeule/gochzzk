package chzzk

import "context"

// GetChzzkUserInfo is a convenience helper equivalent to GetChannels for one channel.
// It uses client authentication and returns the first matching channel, if any.
func GetChzzkUserInfo(ctx context.Context, channelID, clientID, clientSecret string) (*Channel, error) {
	channels, err := New(WithClientCredentials(clientID, clientSecret), WithUserAgent("gochzzk")).GetChannels(ctx, []string{channelID})
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, nil
	}
	return &channels[0], nil
}

// ChannelLiveStatus is a chatbot-friendly live-state result.
type ChannelLiveStatus struct {
	ChannelID    string
	IsLive       bool
	Live         *Live
	ThumbnailURL string
}

// GetChannelLiveStatus scans the cursor-paginated live list for channelID.
// The public API exposes live broadcasts as a list, not a channel-specific endpoint.
func (c *Client) GetChannelLiveStatus(ctx context.Context, channelID string) (ChannelLiveStatus, error) {
	const pageSize = 20
	next := ""
	for {
		page, err := c.GetLives(ctx, pageSize, next)
		if err != nil {
			return ChannelLiveStatus{ChannelID: channelID}, err
		}
		for i := range page.Data {
			if page.Data[i].ChannelID == channelID {
				live := page.Data[i]
				return ChannelLiveStatus{ChannelID: channelID, IsLive: true, Live: &live, ThumbnailURL: live.LiveThumbnailImageURL}, nil
			}
		}
		if page.Page.Next == "" || page.Page.Next == next {
			break
		}
		next = page.Page.Next
	}
	return ChannelLiveStatus{ChannelID: channelID}, nil
}
