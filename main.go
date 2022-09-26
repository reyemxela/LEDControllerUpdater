package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	APP_NAME       = "LED Controller Updater"
	APP_VERSION    = "v1.1.1"
	TMP_DIR_NAME   = "LEDControllerUpdater"
	API_URL        = "https://api.github.com/repos/wingnut-tech/LEDController/releases"
	CH340_URL      = "https://github.com/reyemxela/LEDControllerUpdater/releases/download/v1.0.0/CH34x_Install_Windows_v3_4.zip"
	ZIP_URL_PREFIX = "https://github.com/wingnut-tech/LEDController/archive/refs/tags/"
)

func main() {
	// yay windows
	if runtime.GOOS == "windows" {
		HideConsoleWindow()
	}

	logLevel := flag.String("log", "fatal", "log level: (fatal)/info/debug")
	flag.Parse()

	lvl, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Error(err.Error())
	} else {
		logrus.SetLevel(lvl)
	}

	a := CreateApp()

	a.customSection.Hide()

	go func() {
		a.UpdateReleases()
		a.CheckLibraries()
		a.CheckCore()
		a.GetPorts()
	}()

	a.mainWindow.SetFixedSize(true)
	a.ResizeMainWindow()
	a.mainWindow.CenterOnScreen()

	go func() {
		time.Sleep(1 * time.Second)
		a.CleanOldVersions()
		a.CheckForUpdate()
	}()

	a.mainWindow.ShowAndRun()
}
