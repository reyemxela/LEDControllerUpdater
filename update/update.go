package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/reyemxela/LEDControllerUpdater/releases"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/utils"
)

const (
	APP_API_URL            = "https://api.github.com/repos/reyemxela/LEDControllerUpdater/releases"
	APP_RELEASE_URL_PREFIX = "https://github.com/reyemxela/LEDControllerUpdater/releases"
)

func CheckForUpdate(s *state.State) (bool, string) {
	releases, err := releases.ParseReleases(APP_API_URL)
	if err != nil || len(releases) < 1 {
		return false, ""
	}

	latest := releases[0].Name
	if latest > state.APP_VERSION {
		return true, latest
	}
	return false, ""
}

func UpdateApp(ver string, s *state.State, onSuccess func()) error {
	en, err := GetBinName()
	if err != nil {
		return err
	}
	edir := filepath.Dir(en)

	zipPrefix := "LEDControllerUpdater"
	switch s.AppType {
	case "CLI":
		zipPrefix += "CLI"
	case "GUI":
		if runtime.GOOS == "darwin" {
			return fmt.Errorf("unable to update GUI mac apps")
		}
	default:
		return fmt.Errorf("unknown app type")
	}

	var zipName string
	switch runtime.GOOS {
	case "linux":
		zipName = zipPrefix + "_linux.zip"
	case "windows":
		zipName = zipPrefix + "_windows.zip"
	case "darwin":
		zipName = zipPrefix + "_mac.zip"
	default:
		return fmt.Errorf("unsupported OS")
	}

	s.SetStatus("Updating app...")

	url := fmt.Sprintf("%s/%s/%s/%s", APP_RELEASE_URL_PREFIX, "download", ver, zipName)
	zipFile := filepath.Join(s.TmpDir, zipName)
	err = utils.DownloadFile(zipFile, url)
	if err != nil {
		return err
	}

	bakPath := en + ".bak"
	err = os.Rename(en, bakPath)
	if err != nil {
		return err
	}

	fileNames, err := utils.UnzipFile(zipFile, edir)
	if err != nil {
		return err
	}

	if fileNames[0] != en {
		err = os.Rename(fileNames[0], en)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(en)
	cmd.Start()
	if onSuccess != nil {
		onSuccess()
	}
	os.Exit(0)

	return nil
}

func GetBinName() (string, error) {
	en, err := os.Executable()
	if err != nil {
		return "", err
	}

	en, err = filepath.EvalSymlinks(en)
	if err != nil {
		return "", err
	}

	return en, nil
}

func CleanOldVersions(s *state.State) {
	en, err := GetBinName()
	if err != nil {
		return
	}

	bakPaths := []string{
		en + ".bak",
		filepath.Join(s.TmpDir, filepath.Base(en+".bak")),
	}

	for _, p := range bakPaths {
		os.Remove(p)
	}
}
