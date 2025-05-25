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
	"sync"
	"time"

	"github.com/naruoshin/EnRaiJin/pkg/middleware"
	"github.com/naruoshin/EnRaiJin/pkg/proxy/v2"
)


type Proxmania struct{}

var (
	Version string = "v1.0-beta.2"
	Author  string = "ENKO"
	ProxySourceURL string = "https://raw.githubusercontent.com/proxifly/free-proxy-list/refs/heads/main/proxies/protocols/socks5/data.txt"

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

func (p Proxmania) Run(mw *middleware.Middleware) error {
	// Setting random proxy
	rp := RandomProxy()
	v2.Proxy.Addr = rp.Proxy
	v2.Apply(mw.Client)
	return nil
}

func WriteToFile(proxies []ProxyResult) error {
	file, err := os.OpenFile("proxylist.json", os.O_CREATE|os.O_WRONLY, 0744)
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

func init() {
	// Printing some information about the plugin.
	fmt.Println("\033[1;31m[/!] Plugin 'Proxmania' initializing...\033[0m")
	fmt.Printf("\033[1;31m[-] Version: %s\033[0m\n", Version)
	fmt.Printf("\033[1;31m[-] Author: %s\n\tTwitter: @enkosan_p /x\\ Github: @narukoshin\033[0m\n", Author)

	fmt.Printf("\033[1;32m[-] Proxy data set download in progress...\033[0m\n")

	var err error
	ProxyList, err = Retrieve_ProxyList()
	if err != nil {
		panic(err)
	}
	CheckAndFilter()
	WriteToFile(topProxies)
}

func RandomProxy() ProxyResult {
	rand.Shuffle(len(topProxies), func(i, j int) {
		topProxies[i], topProxies[j] = topProxies[j], topProxies[i]
	})
	return topProxies[0]
}

var mu sync.Mutex

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

	// Saving only top 30 proxies with the best response time
	for i := 0; i < len(AliveProxies) && i < 30; i++ {
		topProxies = append(topProxies, AliveProxies[i])
	}
	// Cleaning garabage that we don't need anymore
	AliveProxies = nil
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

// Retrieving a proxy list from the Github
func Retrieve_ProxyList() ([]string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	// If the proxy is present in the config.yml file,
	// Then the proxy list will be also retrieved using a proxy.
	// 				Anonym1ty Gangz
	if v2.IsProxy() { v2.Apply(client) }
	resp, err := client.Get(ProxySourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var proxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}
	return proxies, scanner.Err()
}