package helper

import "strings"

// SanitizeEmailField removes CR and LF characters to prevent email header injection
func SanitizeEmailField(input string) string {
	sanitized := strings.ReplaceAll(input, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	return sanitized
}
