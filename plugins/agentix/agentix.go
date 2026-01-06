package main

import (
	"math/rand"
	"net/http"
	"github.com/narukoshin/EnRaiJin/v2/pkg/middleware"
)

type Agentix struct{}

var (
	Version string = "v2.0.0"
	Author  string = "Naru Koshin"

	_ middleware.Plugin = Agentix{}
	Plugin = Agentix{}

	DefaultUserAgents = []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:78.0) Gecko/20100101 Firefox/78.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15A372 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/91.0.864.59 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 11; Pixel 4 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 14_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
    }
)

// RandomUserAgent returns a random User-Agent string from the list of pre-defined User-Agents.
func RandomUserAgent() string {
	return DefaultUserAgents[rand.Intn(len(DefaultUserAgents))]
}

// Middleware returns a middleware that sets a random User-Agent header for every HTTP request.
// It wraps the given next RoundTripper and sets the User-Agent header before calling it.
func (p Agentix) Middleware() middleware.ClientMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return middleware.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", RandomUserAgent())
			return next.RoundTrip(req)
		})
	}
}