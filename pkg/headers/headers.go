package headers

import (
	"strings"
	
	"EnRaiJin/pkg/config"
	"EnRaiJin/pkg/structs"
)

func Is() bool {
	return len(config.YAMLConfig.H) != 0
}

func Get() []structs.YAMLHeaders {
	return config.YAMLConfig.H
}

func Find(name string) string {
	for _, h := range Get() {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}