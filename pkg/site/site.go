package site

import (
	"custom-bruteforce/pkg/config"
	"custom-bruteforce/pkg/structs"
	"errors"
	"strings"
)

var (
	Host 	string	= config.YAMLConfig.S.Host
	Method 	string  = config.YAMLConfig.S.Method
	Fields  []structs.YAMLFields = config.YAMLConfig.F
)

// Error message if the request method in the config is incorrect
var ErrInvalidMethod = errors.New("please specify a valid request method")

// All request methods that are allowed to use
var Methods_Allowed []string = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}

// Verifying if the request method is correct
func Verify_Method() error {
	for _, value := range Methods_Allowed {
		if ok := strings.EqualFold(Method, value); ok {
			return nil
		}
	}
	return ErrInvalidMethod
}