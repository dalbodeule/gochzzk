package chatbot

import (
	"testing"

	api "github.com/dalbodeule/gochzzk/src/api"
)

func TestOfficialSessionEventModels(t *testing.T) {
	chat, err := DecodeEvent[api.ChatEvent](api.SessionEvent{Data: []byte(`{"channelId":"c","senderChannelId":"u","chatChannelId":"chat","profile":{"nickname":"nick","badges":[{"badgeNo":1,"badgeId":"b","imageUrl":"https://badge"}],"verifiedMark":true,"userRoleCode":"streamer"},"content":"hello","emojis":{"e":"https://emoji"},"messageTime":123,"eventSentAt":"2026-07-26T03:14:18.629843820"}`)})
	if err != nil || chat.ChatChannelID != "chat" || chat.Profile.Nickname != "nick" || len(chat.Profile.Badges) != 1 || chat.Emojis["e"] == "" || chat.MessageTime != 123 || chat.EffectiveUserRoleCode() != "streamer" {
		t.Fatalf("unexpected chat: %+v, err=%v", chat, err)
	}

	donation, err := DecodeEvent[api.DonationEvent](api.SessionEvent{Data: []byte(`{"donationType":"CHAT","channelId":"c","donatorChannelId":"u","donatorNickname":"nick","payAmount":1000,"donationText":"thanks","emojis":{"e":"https://emoji"},"eventSentAt":"2026-07-26T03:14:28.690447802"}`)})
	if err != nil || donation.DonationType != "CHAT" || donation.PayAmount.String() != "1000" || donation.Emojis["e"] == "" {
		t.Fatalf("unexpected donation: %+v, err=%v", donation, err)
	}

	subscription, err := DecodeEvent[api.SubscriptionEvent](api.SessionEvent{Data: []byte(`{"channelId":"c","subscriberChannelId":"u","subscriberNickname":"n","tierNo":2,"tierName":"gold","month":3}`)})
	if err != nil || subscription.TierNo != 2 || subscription.Month != 3 {
		t.Fatalf("unexpected subscription: %+v, err=%v", subscription, err)
	}
	system, err := DecodeEvent[api.SystemEvent](api.SessionEvent{Data: []byte(`{"type":"subscribed","data":{"eventType":"CHAT","channelId":"c"}}`)})
	if err != nil || system.Data.EventType != "CHAT" || system.Data.ChannelID != "c" {
		t.Fatalf("unexpected system event: %+v, err=%v", system, err)
	}
}
