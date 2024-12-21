package app

import (
	"custom-bruteforce/pkg/bruteforce"
	"custom-bruteforce/pkg/config"
	"custom-bruteforce/pkg/email"
	"custom-bruteforce/pkg/site"
	"custom-bruteforce/pkg/updater"
	"fmt"
	"os"
)

const Version string = "v2.5.2-dev.3"

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
		fmt.Printf("error: %v\n", err)
		return
	}
	// verifying the request method
	if err := site.Verify_Method(); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	// testing email server
	{
		conn, client, err := email.Test_Connection()
		if err != nil {
			fmt.Printf("error: email: %v\n", err)
			return
		}
		// closing the connection after the test is done
		if conn != nil && client != nil {
			client.Close()
			conn.Close()
		}
	}
	// starting a bruteforce attack
	err := bruteforce.Start()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
