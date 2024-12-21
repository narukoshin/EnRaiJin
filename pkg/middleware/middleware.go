package middleware

import (
	"custom-bruteforce/pkg/config"
	"fmt"
	"net/http"
	"os"
	"plugin"
)

var (
	PluginFile string = config.YAMLConfig.B.Plugin
)

type Middleware struct {
	Client *http.Client
	Request *http.Request
}

type Plugin interface {
	Run(*Middleware) error
}

func (m *Middleware) Do() error {
	if PluginFile != "" {
		if _, err := os.Stat(PluginFile); os.IsNotExist(err) {
			return fmt.Errorf("plugin '%s' is not found, please check the path", PluginFile)
		}
		p, err := plugin.Open(PluginFile)
		if err != nil {
			return fmt.Errorf("failed to open plugin: %w", err)
		}
		symbol, err := p.Lookup("Plugin")
		if err != nil {
			return fmt.Errorf("failed to find 'Plugin' symbol: %w", err)
		}
		plugin, ok := symbol.(Plugin)
		if !ok {
			return fmt.Errorf(
				"plugin does not implement the required 'Plugin' interface; found type: %T",
				symbol,
			)
		}
		err = plugin.Run(m)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
