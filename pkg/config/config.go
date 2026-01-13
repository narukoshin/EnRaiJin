package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/creasty/defaults"
	"github.com/narukoshin/EnRaiJin/v2/pkg/structs"
)

// the default config name that we will load.
const DefaultConfigFile string = "config.yml"

var (
	YAMLConfig structs.YAMLConfig
	// Handling errors
	CError error = nil
	// Error message if the config file is not found
	ErrConfigNotFound = "configuration file '%s' could not be found, please make sure that it exists and is readable"
	// Error message if the config file is empty
	ErrConfigIsEmpty = "configuration file '%s' is empty, please make sure that it contains valid configuration"
)

// ParseConfig parses the given YAML config file and unmarshals it into the given config interface.
// It returns an error if the YAML config file is invalid or if the config interface is invalid.
func ParseConfig(yml []byte, config interface{}) error {
	return yaml.Unmarshal(yml, config)
}

// MergeConfig parses the default config file and merges it with the imported config file.
// The imported config file will overwrite the default config file.
// If the imported config file is empty, it will be ignored.
// If the imported config file doesn't exist, an error will be returned.
// If the imported config file is not a string or a []any, an error will be returned.
func MergeConfig(config any) error {
	var err error
	// Parsing default config file
	yml := Load_Config(DefaultConfigFile)
	err = ParseConfig(yml, config)
	if err != nil {
		return err
	}
	// Parsing the imported config file
	// Checking if Import interface is not nil
	if YAMLConfig.Import != nil {
		// checking is "import" string or interface
		switch YAMLConfig.Import.(type) {
		case string:
			{
				yml := Load_Config(YAMLConfig.Import.(string))
				err = ParseConfig(yml, config)
				if err != nil {
					return err
				}
			}
		case []any:
			{
				for _, name := range YAMLConfig.Import.([]any) {
					yml := Load_Config(name.(string))
					err = ParseConfig(yml, config)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Initializes the default configuration by setting default values from the "default" tag in structs.
// It also merges all imported config files into the YAMLConfig struct.
// If any error occurs during initialization, the error is stored in the CError variable.
func init() {
	// Setting default values from "default" tag in structs
	err := defaults.Set(&YAMLConfig)
	if err != nil {
		CError = err
		return
	}
	// Merging config files
	err = MergeConfig(&YAMLConfig)
	if err != nil {
		CError = err
	}
}

// Load_Config loads a YAML config file into a byte array.
// It checks if the config file exists, is readable, and is not empty.
// If any error occurs during loading, the error is stored in the CError variable.
// If the config file is empty or doesn't exist, an error will be returned.
// If the config file exists and is not empty, the contents of the file will be returned as a byte array.
func Load_Config(file_name string) []byte {
	// Checking if the config file exists
	if _, err := os.Stat(file_name); err != nil {
		CError = fmt.Errorf(ErrConfigNotFound, file_name)
		return nil
	}
	// Reading config file
	yml, err := os.ReadFile(file_name)
	if err != nil {
		CError = err
		return nil
	}
	// Checking if the config file is not empty
	if len(yml) == 0 {
		CError = fmt.Errorf(ErrConfigIsEmpty, file_name)
		return nil
	}
	return yml
}

// HasError returns the error that occurred during the initialization of the default configuration.
// If no error occurred, it returns nil.
func HasError() error {
	return CError
}
