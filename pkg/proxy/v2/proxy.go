package v2

import (
	"time"
	"errors"
	"net/url"
	"net/http"
	"crypto/tls"
	"fmt"

	"custom-bruteforce/pkg/structs"
	"custom-bruteforce/pkg/config"
	"custom-bruteforce/pkg/headers"
)


var (
	Proxy structs.YAMLProxy = config.YAMLConfig.P
	IgnoreTLS bool			=  config.YAMLConfig.S.IgnoreTLS
	Timeout time.Duration
	VerifyUrl string 		= config.YAMLConfig.P.VerifyUrl

	ErrTimeoutConversion = errors.New("timeout conversion from string to time.Duration failed")
	ErrInvalidProxyURL 	 = errors.New("invalid proxy URL")
)


func init() {
	if IsProxy() {
		timeout, err := time.ParseDuration(Proxy.Timeout)
		if err != nil {
			config.CError = ErrTimeoutConversion
			return
		}
		Timeout = timeout
		if VerifyUrl == "" { VerifyUrl = "http://httpbin.org/ip" }
		if err = VerifyProxyConnection(); err != nil {
			config.CError = fmt.Errorf("proxy setup failed: %w", err)
			return
		}
	}
}

func VerifyProxyConnection() error {
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

    return nil
}

func Apply(client *http.Client) error {
	url, err := url.Parse(Proxy.Url)
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

func IsProxy() bool {
	return Proxy != (structs.YAMLProxy{})
}