package common

import (
	"net/url"
	"strings"
)

// FindPort returns the port number associated with the given URL scheme.
// If the URL scheme is "https", it returns 443, otherwise it returns 80.
func FindPort(url *url.URL) int {
	if url.Scheme == "https" {
		return 443
	}
	return 80
}

// IsHTTPURL checks if the given string is a valid HTTP or HTTPS URL.
// It does this by checking if the string starts with either "http://" or "https://".
func IsHTTPURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
