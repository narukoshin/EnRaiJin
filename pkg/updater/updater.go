package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/narukoshin/EnRaiJin/v2/pkg/proxy"
	"github.com/walle/targz"
)

const (
	allReleases string = "https://api.github.com/repos/narukoshin/enraijin/releases"
	platform    string = runtime.GOOS
	arch        string = runtime.GOARCH
)

type Mode string

const (
	Loud     Mode = "Loud"
	OnUpdate Mode = "OnUpdate"
)

var binaryFileName string

var Latest Release

type Release struct {
	Version    string          `json:"name"`
	Prerelease bool            `json:"prerelease"`
	Assets     []Release_Asset `json:"assets"`
}

type Release_Asset struct {
	Name     string `json:"name"`
	Download string `json:"browser_download_url"`
}

type HasUpdatesToInstall struct {
	LatestVersion  string
	ExecutableName string
	Assets         []Release_Asset
}

// init sets the binaryFileName based on the platform that the code is running on.
func init() {
	switch platform {
	case "windows":
		binaryFileName = "enraijin.exe"
	case "linux":
		binaryFileName = "enraijin"
	case "darwin":
		binaryFileName = "enraijin"
	default:
		return
	}
}

// Get_Release returns the latest release of Enraijin that is not a prerelease.
// If there are no releases that are not prereleases, it returns an error.
// The function sends a GET request to GitHub's API to get the list of all releases.
// If the request is successful, it parses the JSON response and returns the latest non-prerelease.
func Get_Release() (Release, error) {
	client := &http.Client{}
	// Applying proxy settings if they are available
	if proxy.IsProxy() {
		err := proxy.Apply(client)
		if err != nil {
			return Release{}, err
		}
	}
	resp, err := client.Get(allReleases)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return Release{}, err
		}
		releases := []Release{}
		err = json.Unmarshal(body, &releases)
		if err != nil {
			return Release{}, err
		}
		if len(releases) == 0 {
			return Release{}, fmt.Errorf("couldn't get information about latest releases.")
		}
		for _, release := range releases {
			if !release.Prerelease {
				return release, nil
			}
		}
	}
	return Release{}, nil
}

// CheckForUpdate checks if there is a newer version of Enraijin available.
// It takes two parameters: currentVersion which is the current version of Enraijin that is installed,
// and mode which is the mode of the update checker (Loud or OnUpdate).
// If there is a newer version available, it returns a HasUpdatesToInstall struct with the latest version,
// the name of the executable file, and the assets of the latest release.
// If there is no newer version available, it returns an empty HasUpdatesToInstall struct and no error.
// If the update checker is running in Loud mode, it prints a message when it starts checking for updates,
// and when it finishes checking.
func CheckForUpdate(currentVersion string, mode Mode) (HasUpdatesToInstall, error) {
	if mode == Loud {
		fmt.Printf("\033[36m[-] Checking version...\n\033[0m")
	}
	release, err := Get_Release()
	if err != nil {
		return HasUpdatesToInstall{}, err
	}
	latest_v := release.Version

	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		return HasUpdatesToInstall{}, err
	}
	latestVer, err := semver.NewVersion(latest_v)
	if err != nil {
		return HasUpdatesToInstall{}, err
	}
	if currentVer.LessThan(latestVer) {
		executablePath, err := os.Executable()
		if err != nil {
			return HasUpdatesToInstall{}, err
		}
		return HasUpdatesToInstall{
			LatestVersion:  latest_v,
			ExecutableName: filepath.Base(executablePath),
			Assets:         release.Assets,
		}, nil
	}
	if mode == Loud {
		fmt.Printf("\033[39m[+] You already have the latest available version installed...\n\033[0m")
	}
	return HasUpdatesToInstall{}, nil
}

// SelectAsset selects an asset from the list of available assets that matches the current system's platform and architecture.
// If no matching asset is found, an error is returned with a message indicating that the binary is not available for install and suggesting to compile it manually.
func SelectAsset(assets []Release_Asset) (Release_Asset, error) {
	// Changing naming for macos platforms
	var p string = platform
	if p == "darwin" {
		p = "macos"
	}
	goosarch := fmt.Sprintf("%s-%s", p, arch)
	for _, asset := range assets {
		if strings.Contains(asset.Name, goosarch) {
			return asset, nil
		}
	}
	return Release_Asset{}, fmt.Errorf("sorry, but binary for your system (%s) is not available for install. Please compile it manually.", goosarch)
}

// InstallUpdate checks for the latest version of the binary and updates it if necessary.
// It uses the CheckForUpdate function to check if there is a newer version available and
// if so, it downloads the archive, extracts it, and updates the current binary.
// If the update is successful, it removes the old binary and prints a success message.
// If the binary is locked and cannot be updated, it prints a message asking the user to
// manually delete the old binary file.
// If the update fails, it removes the temp directory and prints an error message.
func InstallUpdate(currentVersion string, mode Mode) error {
	if updates, err := CheckForUpdate(currentVersion, mode); err == nil {
		if updates.LatestVersion != "" {
			var updateSuccess bool = false
			asset, err := SelectAsset(updates.Assets)
			if err != nil {
				return err
			}
			fmt.Printf("\033[36m[-] Newest version found: %s...\n\033[0m", updates.LatestVersion)
			// Creating temp folder
			tempDir, err := os.MkdirTemp("", "enraijin-update-*")
			if err != nil {
				return err
			}
			defer func() {
				fmt.Printf("\033[36m[-] Cleaning up...\n\033[0m")
				os.RemoveAll(tempDir)
				if updateSuccess {
					fmt.Printf("\033[36m[-] Wishing good hacking :PP - ENKO\n\033[0m")
				}
			}()
			fmt.Printf("\033[36m[-] Creating a working directory at %s...\n\033[0m", tempDir)
			fmt.Printf("\033[36m[-] Downloading an archive %s...\n\033[0m", asset.Name)
			tempFile, err := os.Create(filepath.Join(tempDir, asset.Name))
			if err != nil {
				return err
			}

			client := &http.Client{}
			// Applying proxy settings if they are available
			if proxy.IsProxy() {
				err = proxy.Apply(client)
				if err != nil {
					return err
				}
			}
			resp, err := client.Get(asset.Download)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to download archive %s, status: %d...", asset.Name, resp.StatusCode)
			}

			_, err = io.Copy(tempFile, resp.Body)
			if err != nil {
				return err
			}

			fmt.Printf("\033[36m[-] Extracting files...\n\033[0m")
			err = targz.Extract(tempFile.Name(), tempDir)
			if err != nil {
				return err
			}

			binaryPath := filepath.Join(tempDir, binaryFileName)
			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				return fmt.Errorf("binary file %s not found, install failed", binaryFileName)
			}

			currDir, err := os.Getwd()
			if err != nil {
				return err
			}

			fmt.Printf("\033[36m[-] Updating binary...\n\033[0m")
			currentBinary := filepath.Join(currDir, updates.ExecutableName)

			if _, err := os.Stat(currentBinary); os.IsNotExist(err) {
				return fmt.Errorf("Meow! It looks like you're trying to run the update with `go run`. Please use the compiled binary instead.")
			}

			currentBinaryBak := filepath.Join(currDir, fmt.Sprintf("%s_%s", updates.LatestVersion, updates.ExecutableName))
			err = os.Rename(currentBinary, currentBinaryBak)
			if err != nil {
				return err
			}
			defer func() {
				if updateSuccess {
					os.Remove(currentBinaryBak)
					if platform == "windows" {
						fmt.Printf("\033[39m[+] Because you are a Windows user and the binary is locked, I kindly ask you to manually delete the %s binary file...\n\033[0m", currentBinaryBak)
					}
				} else {
					os.Rename(currentBinaryBak, currentBinary)
				}
			}()

			fp, err := os.Create(currentBinary)
			if err != nil {
				return err
			}
			defer fp.Close()

			// reading new binary file to copy
			srcFile, err := os.Open(binaryPath)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			wb, err := io.Copy(fp, srcFile)
			if err != nil {
				return err
			}

			if wb < 0 {
				return fmt.Errorf("failed to copy binary file, written bytes: %d", wb)
			}
			updateSuccess = true
			fmt.Printf("\033[36m[-] The binary successfuly updated...\n\033[0m")
		}
	}
	return nil
}
