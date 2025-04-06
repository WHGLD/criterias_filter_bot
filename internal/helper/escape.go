package helper

import "strings"

var mdV2SpecialChars = []rune{
	'_', '*', '[', ']', '(', ')', '~', '`',
	'>', '#', '+', '-', '=', '{', '}',
	'.', '!', '\\',
	// '|', - нужен для спойлеров
}

// EscapeMarkdownV2 экранирует специальные символы для MarkdownV2
func EscapeMarkdownV2(text string) string {
	var builder strings.Builder

	for _, r := range text {
		if needsEscape(r) {
			builder.WriteRune('\\')
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

func needsEscape(r rune) bool {
	for _, special := range mdV2SpecialChars {
		if r == special {
			return true
		}
	}
	return false
}
