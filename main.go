package main

import (
	"runtime"
)

const (
	APP_NAME       = "LED Controller Updater"
	TMP_DIR_NAME   = "LEDControllerUpdater"
	API_URL        = "https://api.github.com/repos/wingnut-tech/FT-Night-Radian-LED-Controller/releases"
	CH340_URL      = "https://github.com/reyemxela/LEDControllerUpdater/releases/download/v1.0.0/CH34x_Install_Windows_v3_4.zip"
	ZIP_URL_PREFIX = "https://github.com/wingnut-tech/FT-Night-Radian-LED-Controller/archive/refs/tags/"
	FQBN           = "arduino:avr:nano"
	mainWinWidth   = 400
)

var neededLibraries = []string{"FastLED", "Adafruit BMP280 Library"}

func main() {
	// yay windows
	if runtime.GOOS == "windows" {
		HideConsoleWindow()
	}

	a := CreateApp()

	a.customSection.Hide()

	go a.UpdateReleases()
	go a.GetPorts()
	go a.CheckLibraries()
	go a.CheckCore()

	a.mainWindow.SetFixedSize(true)
	a.ResizeMainWindow()
	a.mainWindow.CenterOnScreen()

	a.mainWindow.ShowAndRun()
}