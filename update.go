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
)

const (
	APP_API_URL            = "https://api.github.com/repos/reyemxela/LEDControllerUpdater/releases"
	APP_RELEASE_URL_PREFIX = "https://github.com/reyemxela/LEDControllerUpdater/releases/"
)

func (a *App) CheckForUpdate() {
	releases, err := ParseReleases(APP_API_URL)
	if err != nil || len(releases) < 1 {
		return
	}

	latest := releases[0].Name
	if latest > APP_VERSION {
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

	var zipName string
	if runtime.GOOS == "linux" {
		zipName = "LEDControllerUpdater_linux.zip"
	} else if runtime.GOOS == "windows" {
		zipName = "LEDControllerUpdater_windows.zip"
	}

	url := fmt.Sprintf("%s/%s/%s/%s", APP_RELEASE_URL_PREFIX, "download", ver, zipName)
	zipFile := filepath.Join(a.tmpPath, zipName)
	err = a.DownloadFile(zipFile, url)
	if err != nil {
		return err
	}

	fileNames, err := a.UnzipFile(zipFile, a.tmpPath)
	if err != nil {
		return err
	}

	err = os.Rename(en, filepath.Join(a.tmpPath, filepath.Base(en+".bak")))
	if err != nil {
		return err
	}

	err = os.Rename(fileNames[0], en)
	if err != nil {
		return err
	}

	cmd := exec.Command(en)
	cmd.Start()
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

	os.Remove(filepath.Join(a.tmpPath, filepath.Base(en+".bak")))
}
