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

// CookieMaxAge returns the maximum age for cookies
// based on the COOKIE_MAX_AGE environment variable (defaults to 30 days)
func CookieMaxAge() int {
	maxAgeStr := os.Getenv("COOKIE_MAX_AGE")
	if maxAgeStr == "" {
		return 30 * 24 * 60 * 60 // default to 30 days
	}
	maxAge, err := strconv.Atoi(maxAgeStr)
	if err != nil {
		return 30 * 24 * 60 * 60 // default to 30 days on parse error
	}
	return maxAge
}
