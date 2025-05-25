package app

import (
	"fmt"
	"os"
	
	"github.com/naruoshin/EnRaiJin/pkg/bruteforce"
	"github.com/naruoshin/EnRaiJin/pkg/config"
	"github.com/naruoshin/EnRaiJin/pkg/mail"
	"github.com/naruoshin/EnRaiJin/pkg/site"
	"github.com/naruoshin/EnRaiJin/pkg/updater"
)

const Version string = "v2.5.3-dev"

func Run() {
	// checking if there's any command used
	{
		if len(os.Args) != 1 {
			command := os.Args[1]
			switch command {
			case "version":
				fmt.Printf("\033[33mBuild version: %s\033[0m\r\n", Version)
				if updates, err := updater.CheckForUpdate(Version, updater.Loud); err == nil {
					if updates.LatestVersion != "" {
						fmt.Printf("\033[31mNewer version available to install: %v\033[0m\n\033[36mUse %v update - to install an update\033[0m\r\n", updates.LatestVersion, updates.ExecutableName)
					}
				} else {
					fmt.Printf("error: updater: %v\n", err)
					return
				}
			case "update":
				err := updater.InstallUpdate(Version, updater.Loud)
				if err != nil {
					fmt.Printf("error: updater: %v\n", err)
					return
				}
			}
			return
		}
		// checking for update if we are running a tool
		if updates, err := updater.CheckForUpdate(Version, updater.OnUpdate); err == nil {
			if updates.LatestVersion != "" {
				fmt.Printf("\033[31m[!] There's a new update available to install, to update run \"%v update\"\r\n\033[0m", updates.ExecutableName)
			}
		}
	}
	// verifying the config
	if err := config.HasError(); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	// verifying the target host uri
	if err := site.Verify_Host(); err != nil {
		fmt.Printf("error: site: %v\n", err)
		return
	}
	// verifying the request method
	if err := site.Verify_Method(); err != nil {
		fmt.Printf("error: site: %v\n", err)
		return
	}
	if err := mail.Test(); err != nil {
		fmt.Printf("error: mail: %v\n", err)
		return
	}
	// starting a bruteforce attack
	err := bruteforce.Start()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
