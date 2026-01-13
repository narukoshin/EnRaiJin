package proxy

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

	ErrTimeoutConversion 	= errors.New("timeout conversion from string to time.Duration failed")
	ErrInvalidProxyURL 	 	= errors.New("invalid proxy URL")
	ProxyContextKey 		= struct{}{}
)

type verifyMethod string

const (
    TCP verifyMethod = "TCP"
	HTTP verifyMethod = "HTTP"
)

// Initializes the proxy configuration by setting the default VerifyUrl and parsing the proxy's timeout.
// If the proxy is enabled, it also verifies the connection to the proxy using the TCP method.
// If any error occurs during the verification, the error is stored in the config.CError variable.
func init() {
	if VerifyUrl == "" { VerifyUrl = "http://httpbin.org/ip" }
	if IsProxy() {
		var err error
		fmt.Printf("\033[32m[~] Checking proxy connection \033[0m")
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
		Timeout, err = time.ParseDuration(Proxy.Timeout)
		if err != nil {
			config.CError = ErrTimeoutConversion
			return
		}
		
		if err = VerifyProxyConnection(TCP); err != nil {
			config.CError = fmt.Errorf("proxy: connection verification failed: %w", err)
			return
		}
		fmt.Print(" \033[32mdone.\033[0m\r\n")
	}
}

// SetTimeout sets the timeout for the proxy connection.
// The timeout should be a string in the format accepted by time.ParseDuration.
// If the timeout is invalid, it returns an error.
func SetTimeout(timeout string) error {
	var err error
	Timeout, err = time.ParseDuration(timeout)
	if err != nil {
		return err
	}
	return nil
}

// VerifyProxyConnection checks if the proxy is alive by attempting to establish a connection to it
// using the specified method. It returns an error if the connection fails.
//
// The method can be either "TCP" or "HTTP". If the method is not supported, it returns an error.
//
// The "TCP" method will attempt to establish a TCP connection to the proxy. The "HTTP" method will
// attempt to send a GET request to the specified VerifyUrl through the proxy. If the request fails or
// the response status code is not 200, it returns an error.
//
// The function is used to verify that the proxy is alive and working correctly before using it in the request chain.
//
// Note that the function does not check if the proxy is working correctly for all methods, only that it is
// alive and can be connected to. If the proxy is not working correctly for a specific method, the request
// will fail when sent through the proxy.
func VerifyProxyConnection(method verifyMethod) error {
	switch method {
	case TCP:
		dialer, err := Dial("")
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


// Apply sets the proxy configuration for the given HTTP client.
// It sets the Proxy field of the client's Transport to a function that returns the proxy URL and nil error.
// The function uses the proxy URL stored in the context of the request if available, otherwise it uses the global proxy URL.
// If the global proxy URL is empty, it does not set the proxy URL.
// It also sets the Timeout field of the client to the global timeout and the InsecureSkipVerify field of the client's TLSClientConfig to the global ignore TLS flag.
// If the client's Transport is nil, it sets it to a new *http.Transport. If the client's Transport is not a *http.Transport, it returns an error.
// If the global proxy URL is invalid, it returns an error.
func Apply(client *http.Client) error {
	if client.Transport == nil {
		client.Transport = &http.Transport{}
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		return fmt.Errorf("transport not *http.Transport")
	}
	tr.Proxy = func(req *http.Request) (*url.URL, error) {
		addr, ok := req.Context().Value(ProxyContextKey).(string)
		if !ok || addr == "" {
			if Proxy.Addr == "" {
				return nil, nil
			}
			addr = Proxy.Addr
		}
		url, err := url.Parse(addr)
		if err != nil {
			return nil, ErrInvalidProxyURL
		}
		return url, nil
	}
	client.Timeout = Timeout
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: IgnoreTLS,
	}
	return nil
}

// Dial returns a proxy.Dialer from the given address.
// If the address is empty, it defaults to the value of Proxy.Addr.
// If the proxy is still empty, it returns an error.
// The returned Dialer is created from the parsed URL and a default Dialer.
// If the parsing fails, it returns an error.
func Dial(addr string) (proxy.Dialer, error) {
	// checking if there is a proxy in the parameter
	if addr == "" {
		addr = Proxy.Addr
	}
	// if the proxy is still empty
	if addr == "" {
		return nil, errors.New("no proxy provided")
	}
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.FromURL(url, &net.Dialer{
		Timeout: Timeout,
	})
	if err != nil {
		return nil, err
	}
	return dialer, nil
}

// IsProxy returns true if the Proxy is not empty, false otherwise.
// It can be used to check if a proxy is set before attempting to use it.
func IsProxy() bool {
	return Proxy != (structs.YAMLProxy{})
}