package v2

import (
	"time"
	"errors"
	"net/url"
	"net/http"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/proxy"

	"EnRaiJin/pkg/structs"
	"EnRaiJin/pkg/config"
	"EnRaiJin/pkg/headers"
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
	if IsProxy() {
		timeout, err := time.ParseDuration(Proxy.Timeout)
		if err != nil {
			config.CError = ErrTimeoutConversion
			return
		}
		Timeout = timeout
		
		if err = VerifyProxyConnection(TCP); err != nil {
			config.CError = fmt.Errorf("proxy setup failed: %w", err)
			return
		}
	}
	if VerifyUrl == "" { VerifyUrl = "http://httpbin.org/ip" }
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
		port := func() int {
			if url.Scheme == "https" {
                return 443
            }
            return 80
		}()
		addr := fmt.Sprintf("%s:%d", url.Host, port)
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