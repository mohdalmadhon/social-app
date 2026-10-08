package validation

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MaxIdentifier       = 75
	MaxPassword         = 75
	MinPassword         = 8
	MinName             = 2
	MaxName             = 15
	MinUsername         = 3
	MaxUsername         = 12
	MaxEmail            = 75
	MinEmail            = 5
	MaxAbout            = 1000
	MaxAboutField       = 200
	MaxPostContent      = 1000
	MaxPostLocation     = 200
	MaxPostTags         = 50
	MaxComment          = 200
	MaxCommentLines     = 10
	MaxChatMessage      = 1000
	MaxChatLines        = 15
	MaxGroupTitle       = 15
	MaxGroupDescription = 200
	MaxEventTitle       = 100
	MaxEventDescription = 1000
	MaxSearch           = 100
	MaxCode             = 6
	MaxIDList           = 100
	MaxJSONBody         = 1 << 20
	MaxOffset           = 1000000
	MaxPageLimit        = 100
	MaxFormMemory       = 10 << 20
	MaxAvatarSize       = 5 << 20
	MaxPostMediaSize    = 50 << 20
)

func RuneLen(value string) int {
	return utf8.RuneCountInString(value)
}

func NormalizeNewlines(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}

func CleanText(value string) string {
	return strings.TrimSpace(NormalizeNewlines(value))
}

func CountLines(value string) int {
	if value == "" {
		return 0
	}
	return strings.Count(value, "\n") + 1
}

func HasControlChars(value string) bool {
	for _, r := range value {
		if r == '\n' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

func ValidateSingleLine(label string, value string, min int, max int) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s cannot contain new lines", label)
	}
	return ValidateText(label, value, min, max)
}

func ValidateText(label string, value string, min int, max int) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s contains invalid characters", label)
	}

	if HasControlChars(value) {
		return fmt.Errorf("%s contains invalid characters", label)
	}

	length := RuneLen(value)

	if min > 0 && length < min {
		if min == 1 {
			return fmt.Errorf("%s cannot be empty", label)
		}
		return fmt.Errorf("%s must be at least %d characters", label, min)
	}

	if length > max {
		return fmt.Errorf("%s cannot be more than %d characters", label, max)
	}

	return nil
}

func ValidateMultiline(label string, value string, min int, max int, maxLines int) error {
	if err := ValidateText(label, value, min, max); err != nil {
		return err
	}

	if maxLines > 0 && CountLines(value) > maxLines {
		return fmt.Errorf("%s cannot be more than %d lines", label, maxLines)
	}

	return nil
}

func ValidateSearch(value string) (string, error) {
	value = strings.TrimSpace(value)

	if err := ValidateSingleLine("search", value, 0, MaxSearch); err != nil {
		return "", err
	}

	return value, nil
}

func ValidateChatMessage(content string) (string, error) {
	content = CleanText(content)

	if err := ValidateMultiline("message", content, 1, MaxChatMessage, MaxChatLines); err != nil {
		return "", err
	}

	return content, nil
}

func ValidateIDList(label string, ids []int) error {
	if len(ids) > MaxIDList {
		return fmt.Errorf("%s cannot contain more than %d users", label, MaxIDList)
	}

	for _, id := range ids {
		if id <= 0 {
			return errors.New("invalid user")
		}
	}

	return nil
}

func ClampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > MaxOffset {
		return MaxOffset
	}
	return offset
}

func ClampLimit(limit int, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > MaxPageLimit {
		return MaxPageLimit
	}
	return limit
}
