package unofficial

import "context"

type FollowContent struct {
	UserIDHash        string            `json:"userIdHash"`
	Nickname          string            `json:"nickname"`
	ProfileImageURL   string            `json:"profileImageUrl"`
	UserRoleCode      string            `json:"userRoleCode"`
	Badge             *Badge            `json:"badge"`
	Title             *Title            `json:"title"`
	VerifiedMark      bool              `json:"verifiedMark"`
	ActivityBadges    []Badge           `json:"activityBadges"`
	StreamingProperty StreamingProperty `json:"streamingProperty"`
}

type Badge struct {
	BadgeNo     *int    `json:"badgeNo"`
	BadgeID     *string `json:"badgeId"`
	ImageURL    *string `json:"imageUrl"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Activated   *bool   `json:"activated"`
}
type Title struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
type StreamingProperty struct {
	Following     *Following    `json:"following"`
	NicknameColor NicknameColor `json:"nicknameColor"`
}
type Following struct {
	FollowDate *string `json:"followDate"`
}
type NicknameColor struct {
	ColorCode string `json:"colorCode"`
}

// GetFollowProfile calls the undocumented profile-card endpoint.
func (c *Client) GetFollowProfile(ctx context.Context, chatChannelID, userID string) (*FollowContent, error) {
	chatID, err := escaped(chatChannelID)
	if err != nil {
		return nil, err
	}
	uid, err := escaped(userID)
	if err != nil {
		return nil, err
	}
	rawURL := c.commBaseURL + "/nng_main/v1/chats/" + chatID + "/users/" + uid + "/profile-card?chatType=STREAMING"
	var profile FollowContent
	found, err := c.get(ctx, rawURL, &profile)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &profile, nil
}

// GetFollowDate returns nil when the profile has no following information.
func (c *Client) GetFollowDate(ctx context.Context, chatChannelID, userID string) (*string, error) {
	profile, err := c.GetFollowProfile(ctx, chatChannelID, userID)
	if err != nil || profile == nil || profile.StreamingProperty.Following == nil {
		return nil, err
	}
	return profile.StreamingProperty.Following.FollowDate, nil
}
