package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

const (
	APP_API_URL            = "https://api.github.com/repos/reyemxela/LEDControllerUpdater/releases"
	APP_RELEASE_URL_PREFIX = "https://github.com/reyemxela/LEDControllerUpdater/releases"
)

func (a *App) CheckForUpdate(testUpdate bool) {
	releases, err := ParseReleases(APP_API_URL)
	if err != nil || len(releases) < 1 {
		return
	}

	if testUpdate {
		logrus.Info("Forcing auto-update")
	}

	latest := releases[0].Name
	if latest > APP_VERSION || testUpdate {
		popup := a.fyneApp.NewWindow("Update")
		var content *fyne.Container

		if runtime.GOOS == "darwin" {
			label := widget.NewLabel("New version available: " + latest + ".\n")
			link, _ := url.Parse(fmt.Sprintf("%s/%s/%s", APP_RELEASE_URL_PREFIX, "tag", latest))
			hyperlink := widget.NewHyperlink("Download", link)

			content = container.NewVBox(
				label,
				container.NewHBox(
					hyperlink,
					layout.NewSpacer(),
				),
				widget.NewButton("OK", func() {
					popup.Close()
				}),
			)
		} else {
			label := widget.NewLabel("New version available: " + latest + ".\n\nWould you like to automatically install the update?")
			content = container.NewVBox(
				label,
				layout.NewSpacer(),
				container.NewGridWithColumns(2,
					widget.NewButton("No", func() {
						popup.Close()
					}),
					widget.NewButton("Yes", func() {
						popup.Close()

						err := a.UpdateApp(latest)
						if err != nil {
							a.NewPopup("Error", err.Error())
						}
					}),
				),
			)
		}

		popup.SetContent(content)
		popup.CenterOnScreen()
		popup.Show()
	}
}

func (a *App) UpdateApp(ver string) error {
	en, err := GetBinName()
	if err != nil {
		return err
	}
	edir := filepath.Dir(en)
	logrus.Debug("Current executable name: ", en)

	var zipName string
	if runtime.GOOS == "linux" {
		zipName = "LEDControllerUpdater_linux.zip"
	} else if runtime.GOOS == "windows" {
		zipName = "LEDControllerUpdater_windows.zip"
	}

	url := fmt.Sprintf("%s/%s/%s/%s", APP_RELEASE_URL_PREFIX, "download", ver, zipName)
	zipFile := filepath.Join(a.tmpPath, zipName)
	logrus.Debug("Downloading ", url, " to ", zipFile)
	err = a.DownloadFile(zipFile, url)
	if err != nil {
		return err
	}

	bakPath := en + ".bak"
	logrus.Debug("Renaming running app to ", bakPath)
	err = os.Rename(en, bakPath)
	if err != nil {
		return err
	}

	fileNames, err := a.UnzipFile(zipFile, edir)
	logrus.Debug("Unzipped files: ", fileNames)
	if err != nil {
		return err
	}

	if fileNames[0] != en {
		logrus.Debug("Renaming ", fileNames[0], " to ", en)
		err = os.Rename(fileNames[0], en)
		if err != nil {
			return err
		}
	}

	logrus.Debug("Starting new version")
	cmd := exec.Command(en)
	cmd.Start()
	logrus.Debug("Exiting old version")
	a.fyneApp.Quit()

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

func (a *App) CleanOldVersions() {
	en, err := GetBinName()
	if err != nil {
		return
	}

	bakPaths := []string{
		en + ".bak",
		filepath.Join(a.tmpPath, filepath.Base(en+".bak")),
	}

	for _, p := range bakPaths {
		logrus.Debug("Attempting to remove any old versions at ", p)
		os.Remove(p)
	}
}
