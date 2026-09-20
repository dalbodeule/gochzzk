package chatbot

import (
	"encoding/json"
	"net/url"
	"testing"

	api "github.com/dalbodeule/gochzzk/src/api"
)

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
