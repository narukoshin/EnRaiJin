package v2

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"

	"github.com/narukoshin/EnRaiJin/v2/pkg/common"
	"github.com/narukoshin/EnRaiJin/v2/pkg/config"
	"github.com/narukoshin/EnRaiJin/v2/pkg/headers"
	"github.com/narukoshin/EnRaiJin/v2/pkg/structs"
)


var (
	Proxy structs.YAMLProxy = config.YAMLConfig.P
	IgnoreTLS bool			=  config.YAMLConfig.S.IgnoreTLS
	Timeout time.Duration
	VerifyUrl string 		= config.YAMLConfig.P.VerifyUrl

	ErrTimeoutConversion = errors.New("timeout conversion from string to time.Duration failed")
	ErrInvalidProxyURL 	 = errors.New("invalid proxy URL")
)

type verifyMethod string

const (
    TCP verifyMethod = "TCP"
	HTTP verifyMethod = "HTTP"
)

func init() {
	if VerifyUrl == "" { VerifyUrl = "http://httpbin.org/ip" }
	if IsProxy() {
		var err error
		Timeout, err = time.ParseDuration(Proxy.Timeout)
		if err != nil {
			config.CError = ErrTimeoutConversion
			return
		}
		
		if err = VerifyProxyConnection(TCP); err != nil {
			// config.CError = fmt.Errorf("proxy: connection verification failed: %w", err)
			// return
		}
	}
}

func SetTimeout(timeout string) error {
	var err error
	Timeout, err = time.ParseDuration(timeout)
	if err != nil {
		return err
	}
	return nil
}

func VerifyProxyConnection(method verifyMethod) error {
	switch method {
	case TCP:
		dialer, err := Dial()
		if err != nil {
			return err
		}
		url, err := url.Parse(VerifyUrl)
		if err != nil {
            return err
        }
		port := common.FindPort(url)
		addr := net.JoinHostPort(url.Host, fmt.Sprintf("%d", port))
		if _, err := dialer.Dial("tcp", addr); err != nil {
			return err
		}
	case HTTP:
		client := &http.Client{}
		Apply(client)
		req, err := http.NewRequest(http.MethodGet, VerifyUrl, nil)
		if err != nil {
			return err
		}
		if headers.Find("User-Agent") != "" {
			req.Header.Set("User-Agent", headers.Find("User-Agent"))
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("proxy test failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("proxy test failed: received HTTP status %d", resp.StatusCode)
		}
	default:
		return fmt.Errorf("verification method %s is not supported", method)
	}
    return nil
}

func Apply(client *http.Client) error {
	url, err := url.Parse(Proxy.Addr)
	if err != nil {
		return ErrInvalidProxyURL
	}
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(url),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: IgnoreTLS,
		},
	}
	client.Timeout = Timeout
	return nil
}

func Dial() (proxy.Dialer, error) {
	url, err := url.Parse(Proxy.Addr)
	if err != nil {
		return nil, err
	}
	return proxy.FromURL(url, nil)
}

func IsProxy() bool {
	return Proxy != (structs.YAMLProxy{})
}