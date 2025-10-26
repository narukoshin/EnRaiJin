package main

import (
	"bufio"
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
	Version string = "v1.1"
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
}

var (
	cfg ProxmaniaConfigStruct
	cfgParams ProxmaniaConfigParams
)

// This function will load the config file and parse custom plugin parameters that are not parsed by the default parser.
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

// Run sets a random proxy from the list of alive proxies and applies it to the client of the Middleware object.
func (p Proxmania) Run(mw *middleware.Middleware) error {
	// Setting random proxy
	rp := RandomProxy()
	v2.Proxy.Addr = rp.Proxy
	v2.Apply(mw.Client)
	return nil
}


// WriteToFile writes a list of ProxyResult to a JSON file named "proxylist.json".
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


// Initializes the Proxmania plugin, loads the config file, prints some information about the plugin, retrieves the proxy list...
// filters and checks the proxies, and writes the top proxies to a file.
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


// RandomProxy returns a random proxy from the list of alive proxies.
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


// CheckAndFilter is a function that filters and checks the proxies retrieved from the proxy list,
// removes non-working proxies, and writes the top proxies to a file.
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
		panic("No working proxies found")
	}
}

// This function will try to connect to the website that shows the IP address using a proxy.
// If the connection is sucessfully, it will return a struct with proxy, status, responseTime and bodyResponse
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

// Downloading proxy list from the public data sets
func fetchProxies(client *http.Client, url string) ([]string, error) {
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
    return proxies, nil
}


// Retrieve_ProxyList retrieves a list of proxies from either a local file or a public data set.
// If the proxy is present in the config.yml file, then the proxy list will be also retrieved using a proxy.
func Retrieve_ProxyList() ([]string, error) {
    var proxies []string
    client := &http.Client{
        Timeout: 5 * time.Second,
    }
    // If the proxy is present in the config.yml file,
    // Then the proxy list will be also retrieved using a proxy.
    //              Anonym1ty Gangz
    if v2.IsProxy() {
        v2.Apply(client)
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
			fmt.Printf("\033[1;32m[-] Proxy data set download finished...\033[0m\n")
			return proxies, nil
		}
		// a method that will check if the url is a valid url
		isProtocolSchemed := func(url string) bool {
			return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
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
			return proxies, nil
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
				} else if isProtocolSchemed(proxy.(string)) {
					// Downloading proxies from the public data set
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
				// Downloading proxies from the public data set
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