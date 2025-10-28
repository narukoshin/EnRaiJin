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

type Plugin interface {
	Run(*Middleware) error
}

// The initial idea is that when middleware is being initialized, it will load plugins first and then run them.
// So it doesn't have to open the plugin file on every execution.

// Variable where loaded plugins will be stored at

var LoadedPlugins []Plugin

func InitializePlugins() error {
	err := LoadPlugins()
	if err != nil {
		return err
	}
	return nil
}

// A function that will load plugins
func LoadPlugins() error {
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
	switch Plugins := Plugins.(type) {
	case string: {
		err := openPlugin(Plugins, &LoadedPlugins)
		if err != nil {
			return err
		}
	}
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

func (m *Middleware) Do() error {
	for _, plugin := range LoadedPlugins {
		err := plugin.Run(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// func (m *Middleware) Do() error {
// 	if PluginFile != "" {
// 		if _, err := os.Stat(PluginFile); os.IsNotExist(err) {
// 			return fmt.Errorf("plugin '%s' is not found, please check the path", PluginFile)
// 		}
// 		p, err := plugin.Open(PluginFile)
// 		if err != nil {
// 			return fmt.Errorf("failed to open plugin: %w", err)
// 		}
// 		symbol, err := p.Lookup("Plugin")
// 		if err != nil {
// 			return fmt.Errorf("failed to find 'Plugin' symbol: %w", err)
// 		}
// 		plugin, ok := symbol.(Plugin)
// 		if !ok {
// 			return fmt.Errorf(
// 				"plugin does not implement the required 'Plugin' interface; found type: %T",
// 				symbol,
// 			)
// 		}

// 		// test code
// 		plugins = append(plugins, plugin)
// 		for _, plugin := range plugins {
// 			err = plugin.Run(m)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}
// 	return nil
// }
