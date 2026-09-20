package chzzk

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const MaxChatMessageRunes = 100

var ErrInvalidChatMessageUTF8 = errors.New("chzzk: chat message is not valid UTF-8")

// ChatMessageLengthError reports a chat message longer than the API limit.
type ChatMessageLengthError struct {
	Length int
	Limit  int
}

func (e *ChatMessageLengthError) Error() string {
	return fmt.Sprintf("chzzk: chat message is %d characters; maximum is %d", e.Length, e.Limit)
}

// ValidateChatMessage checks the CHZZK 100-character message limit by Unicode rune count.
func ValidateChatMessage(message string) error {
	if !utf8.ValidString(message) {
		return ErrInvalidChatMessageUTF8
	}
	length := utf8.RuneCountInString(message)
	if length > MaxChatMessageRunes {
		return &ChatMessageLengthError{Length: length, Limit: MaxChatMessageRunes}
	}
	return nil
}

// TruncateChatMessage returns at most MaxChatMessageRunes Unicode characters.
// It preserves valid UTF-8 and does not split a multi-byte character.
func TruncateChatMessage(message string) string {
	runes := []rune(message)
	if len(runes) <= MaxChatMessageRunes {
		return message
	}
	return string(runes[:MaxChatMessageRunes])
}
