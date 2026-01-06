package middleware

import (
	"fmt"
	"net/http"
	"os"
	"plugin"
	
	"github.com/narukoshin/EnRaiJin/v2/pkg/config"
)

var (
	// PluginFile string = config.YAMLConfig.B.Plugin // remove
	// Variable where one or many plugins are stored at
	Plugins any = config.YAMLConfig.B.Plugin
)

type Middleware struct {
	Client *http.Client
	Request *http.Request
}

type ClientMiddleware func(next http.RoundTripper) http.RoundTripper
type RoundTripFunc func(*http.Request) (*http.Response, error)
// RoundTrip calls f(req) and returns its result. It is used to implement
// http.RoundTripper.
//
// The function passed to RoundTrip must follow http.RoundTripper:
// it must not modify the req argument, it must copy the Response before
// returning it to the caller, it must not retain references to the req or
// Response arguments, and it must return (nil, nil) to indicate that the
// request is being retried to the next round tripper in the chain.
//
// The return types of the function must be (*http.Response, error). If f
// returns an error, RoundTrip returns nil, err.
//
// Use RoundTripFunc to create a new http.RoundTripper that delegates
// to the provided function.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type Plugin interface {
	Middleware() ClientMiddleware
}

// The initial idea is that when middleware is being initialized, it will load plugins first and then run them.
// So it doesn't have to open the plugin file on every execution.

// Variable where loaded plugins will be stored at

var LoadedPlugins []Plugin

// InitializePlugins loads all plugins from the given path or paths.
//
// It returns an error if the plugin path is empty, if the plugin is not found, or if the symbol 'Plugin' is not found in the plugin.
//
// It will append the loaded plugins to the LoadedPlugins slice.
func InitializePlugins() error {
	err := LoadPlugins()
	if err != nil {
		return err
	}
	return nil
}

// LoadPlugins loads plugins from the given path or paths.
//
// If the given paths are strings, it will load a single plugin from the given path.
// If the given paths are slices of any, it will load plugins from the given paths.
//
// It returns an error if the plugin path is empty, if the plugin is not found, or if the symbol 'Plugin' is not found in the plugin.
//
// The function will append the loaded plugins to the LoadedPlugins slice.
func LoadPlugins() error {
	// Function to open a plugin
	openPlugin := func (p string, PluginList *[]Plugin) error {
		// debug note, remove later
		fmt.Println("DEBUG: Loading plugin: " + p)
		if p == "" {
			return fmt.Errorf("path to the plugin is not specified")
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("plugin '%s' is not found, please check the path", p)
		}
		pf, err := plugin.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open plugin: %w", err)
		}
		symbol, err := pf.Lookup("Plugin")
		if err != nil {
			return fmt.Errorf("failed to find 'Plugin' symbol: %w", err)
		}
		plugin, ok := symbol.(Plugin)
		if !ok {
			return fmt.Errorf("failed to cast 'Plugin' symbol to Plugin interface: %w", err)
		}
		*PluginList = append(*PluginList, plugin)
		return nil
	}
	// Switch to determine the type of the Plugins variable
	switch Plugins := Plugins.(type) {
	// if there is only one plugin specified
	case string: {
		err := openPlugin(Plugins, &LoadedPlugins)
		if err != nil {
			return err
		}
	}
	// if there are multiple plugins specified
	case []any: {
		for _, plugin := range Plugins {
			err := openPlugin(plugin.(string), &LoadedPlugins)
			if err != nil {
				return err
			}
		}
	}
	}
	return nil
}

// Do executes the HTTP request with the given request and client.
// If the client is nil, it creates a new client.
// If the transport of the client is nil, it uses the default transport.
// After that, it applies the middlewares from the loaded plugins and sets the transport of the client.
// Finally, it executes the request and returns the response.
// The original transport of the client is restored after the request is finished.
func (m *Middleware) Do() (*http.Response, error) {
	// If the client is nil, create a new client
	if m.Client == nil {
		m.Client = &http.Client{}
	}
	// Save the original transport
	ogTr := m.Client.Transport
	// If the transport is nil, use the default transport
	if ogTr == nil {
		ogTr = http.DefaultTransport
	}
	// Apply the middlewares
	tr := ogTr
	// Executing the middlewares of the plugins
	for _, plugin := range LoadedPlugins {
		tr = plugin.Middleware()(tr)
	}
	// Set the transport
	m.Client.Transport = tr
	defer func(){
		// Restore the original transport
		m.Client.Transport = ogTr
	}()
	// Execute the request
	return m.Client.Do(m.Request)
}
