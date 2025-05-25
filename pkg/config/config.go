package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	
	"github.com/naruoshin/EnRaiJin/pkg/structs"
)

// the default config name that we will load.
const YAMLFile string = "config.yml"

var (
	YAMLConfig structs.YAMLConfig

	// Handling errors
	CError error = nil

	// Error message if the config file is not found
	ErrConfigNotFound = errors.New("config file not found, please create it to use this tool")

	// Error message if the config file is empty
	ErrConfigIsEmpty = errors.New("config file is empty")

	// Error message if the include file doesn't exist
	// Almost duplication of ErrConfigNotFound
	ErrIncludeNotFound = errors.New("one or more include files not found")
)

func init() {
	yml := load_file(YAMLFile)
	err := yaml.Unmarshal(yml, &YAMLConfig)
	if err != nil {
		CError = err
		return
	}

	// if `import` option is not empty, then importing the file from the option
	// if the `import` option is empty, we will use config.yml that is loaded above.
	if len(YAMLConfig.Import) != 0 {
		yml = load_file(YAMLConfig.Import)
		err := yaml.Unmarshal(yml, &YAMLConfig)
		if err != nil {
			CError = err
			return
		}
	}
	// Including config by parts
	if len(YAMLConfig.Include) != 0 {
		// Because include option is an array
		// We need to iterate through it
		for _, inc := range YAMLConfig.Include {
			// Trying to read the file and load it.
			yml = load_file(inc)
			// If any of include files doesn't exist
			// Returning an error message
			if CError == ErrConfigNotFound {
				CError = ErrIncludeNotFound
			}
			err = yaml.Unmarshal(yml, &YAMLConfig)
			if err != nil {
				CError = err
				return
			}
		}
	}
}

func load_file(file_name string) []byte {
	// Checking if the config file exists
	if _, err := os.Stat(file_name); err != nil {
		CError = ErrConfigNotFound
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
		CError = ErrConfigIsEmpty
		return nil
	}
	return yml
}

// Checking for errors in the app.Run function
func HasError() error {
	return CError
}
