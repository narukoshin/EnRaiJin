package site

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/narukoshin/EnRaiJin/v2/pkg/common"
	"github.com/narukoshin/EnRaiJin/v2/pkg/config"
	"github.com/narukoshin/EnRaiJin/v2/pkg/proxy"
	"github.com/narukoshin/EnRaiJin/v2/pkg/structs"
)

var (
	Host   string               = config.YAMLConfig.S.Host
	Method string               = config.YAMLConfig.S.Method
	Fields []structs.YAMLFields = config.YAMLConfig.F

	// Error message if the request method in the config is incorrect
	ErrInvalidMethod = errors.New("please specify a valid request method")
	ErrDeadHost      = errors.New("looks that the host is not alive, check your config again")

	// All request methods that are allowed to use
	Methods_Allowed []string = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
)

// Verify_Method checks if the request method specified in the config is valid.
// It checks if the method is one of the following: GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH.
// If the method is not valid, it returns an error message.
func Verify_Method() error {
	for _, value := range Methods_Allowed {
		if ok := strings.EqualFold(Method, value); ok {
			return nil
		}
	}
	return ErrInvalidMethod
}

// Verify_Host checks if the host specified in the config is valid and alive.
// It first checks if the host is set correctly with the scheme, then gets the host port from the scheme.
// Finally, it checks if the host is alive by attempting to establish a TCP connection to the host.
// If the host is not valid or alive, it returns an error message.
func Verify_Host() error {
	// checking if the host is set correctly with the scheme
	url, err := url.ParseRequestURI(Host)
	if err != nil {
		return err
	}
	// getting the host port from the scheme that is specified in the config
	port := common.FindPort(url)
	// checking if the host is alive

	if proxy.IsProxy() { // checking with the proxy
		dialer, err := proxy.Dial("")
		if err != nil {
			return err
		}
		if _, err := dialer.Dial("tcp", net.JoinHostPort(url.Host, fmt.Sprintf("%d", port))); err != nil {
			return err
		}
	} else { // checking without the proxy
		if _, err = net.DialTimeout("tcp", net.JoinHostPort(url.Host, fmt.Sprintf("%d", port)), 3*time.Second); err != nil {
			return ErrDeadHost
		}
	}
	return nil
}
