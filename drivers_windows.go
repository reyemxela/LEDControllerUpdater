package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gonutz/ide/w32"
)

// ch340 drivers are only needed on windows, because windows is so special.
// (for more windows rant, see below)
func (a *App) InstallCH340() {
	exeFile := filepath.Join(a.tmpPath, "CH34x_Install_Windows_v3_4.EXE")
	zipFile := filepath.Join(a.tmpPath, "ch340.zip")

	if _, err := os.Stat(exeFile); err != nil {
		if _, err := os.Stat(zipFile); err != nil {
			err := a.DownloadFile(zipFile, CH340_URL)
			if err != nil {
				a.NewPopup("Error", err.Error())
				return
			}
		}

		_, err = a.UnzipFile(zipFile, a.tmpPath)
		if err != nil {
			a.NewPopup("Error", err.Error())
			return
		}
	}

	a.NewPopup("Info", "Make sure to plug in the controller before hitting the Install button")
	time.Sleep(time.Second * 2)

	cmd := exec.Command(filepath.Join(a.tmpPath, "CH34x_Install_Windows_v3_4.EXE"))
	err := cmd.Start()
	if err != nil {
		a.NewPopup("Error", err.Error())
	}
}

// yet another of the many reasons windows is garbage and I hate it.
// I have to build the windows binary as a console app (no windowsgui flag),
// and then, using this function, manually hide the console window that spawns.
// if I don't, when trying to compile a custom firmware, dozens of console windows
// will spam across the screen as it compiles, flashing and instantly closing again.
func HideConsoleWindow() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_HIDE)
		}
	}
}
