package helper

import (
	"net/http"
	"os"
	"strconv"
)

// GetCookieSecure returns whether cookies should have the Secure flag set
// based on the COOKIE_SECURE environment variable (defaults to true)
func GetCookieSecure() bool {
	secureStr := os.Getenv("COOKIE_SECURE")
	if secureStr == "" {
		return true // default to true for security
	}
	secure, err := strconv.ParseBool(secureStr)
	if err != nil {
		return true // default to true on parse error
	}
	return secure
}

// GetCookieSameSite returns the SameSite mode for cookies
// Always returns SameSiteNoneMode for cross-origin support
func GetCookieSameSite() http.SameSite {
	return http.SameSiteNoneMode
}
