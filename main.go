package main

import (
	"runtime"
)

const (
	APP_NAME       = "LED Controller Updater"
	APP_VERSION    = "v1.1.0"
	TMP_DIR_NAME   = "LEDControllerUpdater"
	API_URL        = "https://api.github.com/repos/wingnut-tech/LEDController/releases"
	CH340_URL      = "https://github.com/reyemxela/LEDControllerUpdater/releases/download/v1.0.0/CH34x_Install_Windows_v3_4.zip"
	ZIP_URL_PREFIX = "https://github.com/wingnut-tech/LEDController/archive/refs/tags/"
	mainWinWidth   = 400
)

func main() {
	// yay windows
	if runtime.GOOS == "windows" {
		HideConsoleWindow()
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

	a.mainWindow.ShowAndRun()
}
