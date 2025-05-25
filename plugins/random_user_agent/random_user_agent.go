package main

import (
	"math/rand"
	"time"
	
	"github.com/narukoshin/EnRaiJin/v2/pkg/middleware"
)

type UserAgentPlugin struct{}

var (
	_ middleware.Plugin = UserAgentPlugin{}
	Plugin = UserAgentPlugin{}

	UserAgents = []string{
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

func init(){
	rand.Seed(time.Now().UnixNano())
}

func RandomUserAgent() string {
	rand.Shuffle(len(UserAgents), func(i, j int) {
		UserAgents[i], UserAgents[j] = UserAgents[j], UserAgents[i]
	})
	return UserAgents[0]
}

func (p UserAgentPlugin) Run(m *middleware.Middleware) error {
	m.Request.Header.Set("User-Agent", RandomUserAgent())
	return nil
}