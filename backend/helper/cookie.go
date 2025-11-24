package helper

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var cookieSecureOnce sync.Once
var isCookieSecure bool

// Initialize Cookie Secure
func initCookieSecure() {
	secureStr := os.Getenv("COOKIE_SECURE")

	// Default to true for security if variable is unset or empty
	if secureStr == "" {
		isCookieSecure = true
		return
	}

	secure, err := strconv.ParseBool(secureStr)
	// Default to true on parse error
	if err != nil {
		isCookieSecure = true
		slog.Warn("Invalid value for COOKIE_SECURE. Defaulting to true", "error", err)
		return
	}
	isCookieSecure = secure
}

// GetCookieSecure returns whether cookies should have the Secure flag set
// based on the COOKIE_SECURE environment variable (defaults to true)
func GetCookieSecure() bool {
	cookieSecureOnce.Do(initCookieSecure)
	return isCookieSecure
}

// GetCookieName returns the canonical name for a cookie based on the base name and the secure flag
func GetCookieName(baseName string) (string, error) {
	if baseName == "" {
		return "", errors.New("cookie name cannot be empty")
	}

	secure := GetCookieSecure()
	var canonicalName string

	if secure {
		canonicalName = "__Secure-" + baseName
	} else {
		canonicalName = baseName
	}

	return canonicalName, nil
}

func GetRefreshCookieName() string {
	name, err := GetCookieName("refresh_token")
	if err != nil {
		slog.Error("Failed to get refresh cookie name", "error", err)
		return ""
	}
	return name
}

// GetCookieSameSite returns the SameSite mode for cookies
// Always returns SameSiteNoneMode for cross-origin support
func GetCookieSameSite() http.SameSite {
	return http.SameSiteNoneMode
}
