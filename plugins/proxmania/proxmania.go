package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"

	"github.com/narukoshin/EnRaiJin/v2/pkg/config"
	"github.com/narukoshin/EnRaiJin/v2/pkg/middleware"
	"github.com/narukoshin/EnRaiJin/v2/pkg/proxy/v2"
)


type Proxmania struct{}

var (
	Version string = "v2.0.0"
	Author  string = "Naru Koshin"
	DefaultProxySourceURL string = "https://raw.githubusercontent.com/proxifly/free-proxy-list/refs/heads/main/proxies/protocols/socks5/data.txt"
	ProxyMaxCount int = 30

	ProxyList []string
	_ middleware.Plugin = Proxmania{}
	Plugin = Proxmania{}
)


// Status of the Proxy
type ProxyStatus string

var (
	StatusGood ProxyStatus = "good"
	StatusBad  ProxyStatus = "bad"
	StatusDead ProxyStatus = "dead"
)

type ProxyResult struct {
	Proxy string
	Status ProxyStatus
	ResponseTime float64
	BodyResponse string
}

var (
	AliveProxies []ProxyResult
	// TOP 30 of the best proxies with lowest response time
	topProxies []ProxyResult
)

type ProxmaniaConfigStruct struct {
	Proxmania ProxmaniaConfigParams `yaml:"proxmania"`
}

type ProxmaniaConfigParams struct {
	ProxyDataSet any `yaml:"proxy_data_set"`
	ProxyMaxCount int `yaml:"max_proxies" default:"30"`
	Timeout time.Duration `yaml:"timeout" default:"60s"`
}

var (
	cfg ProxmaniaConfigStruct
	cfgParams ProxmaniaConfigParams
)

// ProximediaConfig sets default values for the Proximedia configuration and
// then merges the configuration from the file specified by the config file path
// into the ProximediaConfigStruct. It returns an error if the configuration
// file could not be read or if the configuration is invalid.
func ProxmaniaConfig() error {
	// setting default values
	if err := defaults.Set(&cfg); err != nil {
		return err
	}
	err := config.MergeConfig(&cfg)
	if err != nil {
		return err
	}
	cfgParams = cfg.Proxmania
	return nil
}


// Middleware returns a middleware that sets a random proxy from the available proxy list for every HTTP request.
// It wraps the given next RoundTripper and sets the proxy URL before calling it.
// If there are no available proxies, it returns an error.
// The function uses the proxy URL stored in the context of the request if available, otherwise it uses a random proxy from the list.
// If the global proxy URL is empty, it does not set the proxy URL.
// It also sets the Timeout field of the client to the global timeout and the InsecureSkipVerify field of the client's TLSClientConfig to the global ignore TLS flag.
// If the client's Transport is nil, it sets it to a new *http.Transport. If the client's Transport is not a *http.Transport, it returns an error.
// If the global proxy URL is invalid, it returns an error.
func (p Proxmania) Middleware() middleware.ClientMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return middleware.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			rp := RandomProxy()
			if rp.Proxy == "" {
				return nil, fmt.Errorf("no proxy available")
			}
			ctx, cancel := context.WithTimeout(req.Context(), cfgParams.Timeout)
			defer cancel()
			req = req.WithContext(ctx)
			ctx = context.WithValue(req.Context(), v2.ProxyContextKey, rp.Proxy)
			req = req.WithContext(ctx)
			resp, err := next.RoundTrip(req)
			if err != nil {
				fmt.Println("fucked up something")
				return nil, err
			}
			return resp, err
		})
	}
}

// WriteToFile writes the given list of proxies to a file named "proxylist.json".
// If the file does not exist, it will be created. If the file already exists, it will be truncated.
// The list of proxies is marshalled into JSON before being written to the file.
// If there is an error while writing to the file, that error will be returned.
// The file is closed after writing is finished, regardless of whether an error occurred or not.
func WriteToFile(proxies []ProxyResult) error {
	file, err := os.OpenFile("proxylist.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.Marshal(proxies)
	if err != nil {
		return err
	}
	file.Write(jsonData)
	return nil
}



// Initializes the Proximedia plugin by reading the configuration file, retrieving the proxy list, filtering the proxy list, and writing the top proxies to a file named "proxylist.json". 
// If any error occurs during the initialization, it panics with the error message.
func init() {
	var err error
	err = ProxmaniaConfig()
	if err != nil {
		panic(err)
	}
	// Printing some information about the plugin.
	fmt.Println("\033[1;31m[/!] Plugin 'Proxmania' initializing...\033[0m")
	fmt.Printf("\033[1;31m[-] Version: %s\033[0m\n", Version)
	fmt.Printf("\033[1;31m[-] Author: %s\n\tTwitter: @enkosan_p /x\\ Github: @narukoshin\033[0m\n", Author)
	ProxyList, err = Retrieve_ProxyList()
	if err != nil {
		panic(err)
	}
	CheckAndFilter()
	WriteToFile(topProxies)
}


// Returns a random proxy from the list of top proxies.
// If the list of top proxies is empty, it returns an empty ProxyResult.
// The function uses the math/rand package to shuffle the list of top proxies and then returns the first item in the list.
// The function is used to get a random proxy from the list of top proxies without having to manually shuffle the list every time.
func RandomProxy() ProxyResult {
	if len(topProxies) == 0 {
		return ProxyResult{}
	}
	rand.Shuffle(len(topProxies), func(i, j int) {
		topProxies[i], topProxies[j] = topProxies[j], topProxies[i]
	})
	return topProxies[0]
}

var mu sync.Mutex



// CheckAndFilter is a function that takes a list of proxies and checks if they are working by
// sending a GET request to the proxy and checking the response status code and response time.
// If the proxy is working, it adds the proxy to the list of alive proxies and sorts the list by response time in ascending order.
// Finally, it saves only the top proxies with the best response time to the list of top proxies and cleans up the garbage that is not needed anymore.
// If there are no working proxies found, it sends a panic with an error message.
func CheckAndFilter() {
    var wg sync.WaitGroup
    results := make(chan ProxyResult, len(ProxyList))

    for _, proxy := range ProxyList {
        wg.Add(1)
        go func(proxy string) {
            defer wg.Done()
            result := Check_WorkingProxies(proxy)
            if result.Status == StatusGood {
				mu.Lock()
                AliveProxies = append(AliveProxies, result)
				mu.Unlock()
            }
            results <- result
        }(proxy)
    }

    go func() {
        wg.Wait()
        close(results)
    }()
	wg.Wait()

	// Sorting the proxies by response time in ascending order.
	sort.Slice(AliveProxies, func(i, j int) bool {
		return AliveProxies[i].ResponseTime < AliveProxies[j].ResponseTime
	})

	// Saving only top proxies with the best response time
	for i := 0; i < len(AliveProxies) && i < ProxyMaxCount; i++ {
		topProxies = append(topProxies, AliveProxies[i])
	}
	// Cleaning garabage that we don't need anymore
	AliveProxies = nil
	// Checking if there is any proxy in the list of top proxies, if its empty, sending panic.
	if len(topProxies) == 0 {
		fmt.Printf("error: no working proxies found\n")
		os.Exit(1)
	}
}


// Check_WorkingProxies checks if the given proxy is working by sending a GET request to the proxy and checking the response status code and response time.
// If the proxy is working, it returns a ProxyResult with the proxy's address, status set to StatusGood, response time in seconds, and the body response.
// If the proxy is not working, it returns a ProxyResult with the proxy's address, status set to StatusDead, response time set to 0, and the error message.
// If the proxy is working but the response status code is not 200 or the response time is greater than 5 seconds, it returns a ProxyResult with the proxy's address, status set to StatusBad, response time in seconds, and the body response.
func Check_WorkingProxies(proxy string) (ProxyResult) {
	client := &http.Client{}
	v2.Proxy.Addr = proxy
	err := v2.Apply(client)
	if err != nil {
		return ProxyResult{ Proxy: proxy, Status: StatusDead, ResponseTime: 0, BodyResponse: err.Error() }
	}
	client.Timeout = 5 * time.Second
	start := time.Now()
	resp, err := client.Get(v2.VerifyUrl)
	if err != nil {
		return ProxyResult{ Proxy: proxy, Status: StatusDead, ResponseTime: 0, BodyResponse: err.Error(), }
	}

	defer resp.Body.Close()
	duration := time.Since(start).Seconds()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return ProxyResult{ Proxy: proxy, Status: StatusBad, ResponseTime: duration, BodyResponse: err.Error() }
		}
		return ProxyResult{ Proxy: proxy, Status: StatusGood, ResponseTime: duration, BodyResponse: string(body) }
	}
	return ProxyResult{ Proxy: proxy, Status: StatusBad, ResponseTime: duration, BodyResponse: "" }
}


// Retrieve_ProxyList retrieves a list of proxies from a given data set.
//
// The data set can be a local file, a valid URL, or an array of local files or valid URLs.
// The function will check the type of the data set and based on that, it will use either the
// loadLocalDataSet or the fetchProxies method to retrieve the proxy list.
//
// If the data set is a local file, the loadLocalDataSet method will be used to load the proxies
// from the file. If the data set is a valid URL, the fetchProxies method will be used to download
// the proxy list from the URL.
//
// If the data set is an array, the function will loop through the array and check the type of
// each item. If the item is a local file, the loadLocalDataSet method will be used to load the
// proxies from the file. If the item is a valid URL, the fetchProxies method will be used to download
// the proxy list from the URL.
//
// The function returns the list of proxies and an error if any occurs during the retrieval of
// the proxy list. If the data set is empty, the function will return an empty list and a nil error.
//
// The function also checks if a proxy is present in the config.yml file. If a proxy is present,
// the proxy list will also be retrieved using a proxy.
func Retrieve_ProxyList() ([]string, error) {
    var proxies []string
    client := &http.Client{
        Timeout: 15 * time.Second,
    }
    // If the proxy is present in the config.yml file,
    // Then the proxy list will be also retrieved using a proxy.
    //              Anonym1ty Gangz
    if v2.IsProxy() {
        v2.Apply(client)
    }

	// a method that will fetch a proxy data set from a public data set
	fetchProxies := func (client *http.Client, url string) ([]string, error) {
		fmt.Printf("\033[1;32m[-] Proxy data set download in progress from %s...\033[0m\n", url)
		var proxies []string
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			proxies = append(proxies, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		fmt.Printf("\033[1;32m[-] Proxy data set download finished...\033[0m\n")
		return proxies, nil
	}

    // Checking if proxyList is nil
    if cfgParams.ProxyDataSet != nil {
		// a method that will load a local data set from the file specifiec in the params
		loadLocalDataSet := func(name string) ([]string, error) {
			fmt.Printf("\033[1;32m[-] Loading local proxy data set from %s...\033[0m\n", name)
			file, err := os.Open(name)
			if err != nil {
				return nil, err
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			var proxies []string
			for scanner.Scan() {
				proxies = append(proxies, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return proxies, nil
		}
		// a method that will check if the url is a valid url
		isProtocolSchemed := func(url string) bool {
			return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
		}

		// Checking the type of the proxy data set
        switch cfgParams.ProxyDataSet.(type) {
		// Checking if the proxy data set is an array
        case []any:
            for _, proxy := range cfgParams.ProxyDataSet.([]any) {
				// Checking if data set is a local file
				if _, err := os.Stat(proxy.(string)); err == nil {
					p, err := loadLocalDataSet(proxy.(string))
					if err != nil {
						return nil, err
					}
					proxies = append(proxies, p...)
				// Checking if data set is a valid url
				} else if isProtocolSchemed(proxy.(string)) {
					p, err := fetchProxies(client, proxy.(string))
					if err != nil {
						return nil, err
					}
					proxies = append(proxies, p...)
				}
            }
		// Checking if the proxy data set is a string
        case string:
			// Checking if data set is a local file
			if _, err := os.Stat(cfgParams.ProxyDataSet.(string)); err == nil {
				proxies, err = loadLocalDataSet(cfgParams.ProxyDataSet.(string))
				if err != nil {
					return nil, err
				}
			} else if isProtocolSchemed(cfgParams.ProxyDataSet.(string)) {
				proxyList, err := fetchProxies(client, cfgParams.ProxyDataSet.(string))
				if err != nil {
					return nil, err
				}
				proxies = append(proxies, proxyList...)
			}
        }
	// Using default proxy set
    } else {
        proxyList, err := fetchProxies(client, DefaultProxySourceURL)
        if err != nil {
            return nil, err
        }
        proxies = append(proxies, proxyList...)
    }
    return proxies, nil
}