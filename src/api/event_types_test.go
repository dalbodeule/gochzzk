package chzzk

import (
	"testing"
	"time"
)

func TestParseEventSentAtAsKST(t *testing.T) {
	got, err := ParseEventSentAt("2026-07-26T03:14:18.629843820")
	if err != nil {
		t.Fatal(err)
	}
	_, offset := got.Zone()
	if offset != 9*60*60 || got.Nanosecond() != 629843820 {
		t.Fatalf("unexpected parsed time: %s", got.Format(time.RFC3339Nano))
	}
	wantUTC := time.Date(2026, 7, 25, 18, 14, 18, 629843820, time.UTC)
	if !got.UTC().Equal(wantUTC) {
		t.Fatalf("unexpected UTC conversion: got %s, want %s", got.UTC(), wantUTC)
	}
}

func TestChatEventMessageTimeUTC(t *testing.T) {
	event := ChatEvent{MessageTime: 1_700_000_000_123}
	want := time.Date(2023, 11, 14, 22, 13, 20, 123_000_000, time.UTC)
	got := event.MessageTimeUTC()
	if !got.Equal(want) || got.Location() != time.UTC {
		t.Fatalf("unexpected messageTime: got %s, want %s", got, want)
	}
}
