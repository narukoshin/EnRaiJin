package v1

import (
	"golang.org/x/net/proxy"
	"h12.io/socks"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
	"net"
	
	"github.com/naruoshin/EnRaiJin/pkg/structs"
	"github.com/naruoshin/EnRaiJin/pkg/config"
)

var Proxy structs.YAMLProxy =  config.YAMLConfig.P
var IgnoreTLS bool			=  config.YAMLConfig.S.IgnoreTLS

func dial_socks() *http.Transport {
	dialSocks := socks.Dial(Proxy.Socks)
	if IgnoreTLS {
		return &http.Transport{Dial: dialSocks, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return &http.Transport{Dial: dialSocks}
}

func IsProxy() bool {
	return Proxy.Socks != ""
}

func Drive() *http.Transport {
	if Proxy.Socks != "" {
		return dial_socks()
	}
	return &http.Transport{}
}

func Dialer(timeout time.Duration) (proxy.Dialer, error){
	parsed, err := parse_proxy(Proxy.Socks)
	if err != nil {
		return nil, err
	}
	return proxy.SOCKS5("tcp", parsed.Host, nil, &net.Dialer{Timeout: timeout * time.Second})
}

func parse_proxy(proxy string) (*url.URL, error) {
	return url.ParseRequestURI(proxy)
}