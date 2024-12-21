package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"os"
	"path/filepath"
	"runtime"
	"github.com/Masterminds/semver"
	"github.com/walle/targz"
)

const (
	allReleases 		string = "https://api.github.com/repos/narukoshin/enraijin/releases"
	platform		 	string = runtime.GOOS
	arch				string = runtime.GOARCH
)

type Mode string

const (
	Loud Mode = "Loud"
	OnUpdate Mode = "OnUpdate"
)


var binaryFileName string

var Latest Release

type Release struct {
	Version	   string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Assets []Release_Asset `json:"assets"`
}

type Release_Asset struct {
	Name string `json:"name"`
	Download string `json:"browser_download_url"`
}

type HasUpdatesToInstall struct {
	LatestVersion string
	ExecutableName string
	Assets []Release_Asset
}

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

func Get_Release() (Release, error) {
	resp, err := http.Get(allReleases)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
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
	if currentVer.LessThan(latestVer){
		executablePath, err := os.Executable()
		if err != nil {
			return HasUpdatesToInstall{}, err
		}
		return HasUpdatesToInstall {
			LatestVersion: latest_v,
			ExecutableName: filepath.Base(executablePath),
			Assets: release.Assets,
		}, nil
	}
	if mode == Loud {
		fmt.Printf("\033[39m[+] You already have the latest available version installed...\n\033[0m")
	}
	return HasUpdatesToInstall{}, nil
}


func SelectAsset(assets []Release_Asset) (Release_Asset, error)  {
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
			defer func(){
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

			resp, err := http.Get(asset.Download)
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
			if err != nil {
				return err
			}

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