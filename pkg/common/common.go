package common

import (
	"net/url"
)

func FindPort(url *url.URL) int {
	if url.Scheme == "https" {
		return 443
	}
	return 80
}