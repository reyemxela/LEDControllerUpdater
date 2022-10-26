package main

import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/reyemxela/LEDControllerUpdater/arduino"
	"github.com/reyemxela/LEDControllerUpdater/common"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/update"
	"github.com/reyemxela/LEDControllerUpdater/utils"
)

func main() {
	if runtime.GOOS == "windows" {
		hideConsole()
	}

	ui := &UI{}
	s, err := state.NewState("GUI", ui.setStatus)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ui.state = s

	ui.app = app.New()

	ui.mainWindow = ui.app.NewWindow(state.APP_NAME)
	ui.mainWindow.SetContent(createMainWindow(ui))

	go arduino.WatchPorts(ui.state, func() {
		ui.portList.Options = utils.ListKeys(ui.state.Ports)
		if len(ui.portList.Options) < 1 {
			ui.portList.ClearSelected()
		} else {
			ui.portList.SetSelected(ui.state.CurrentPort)
		}
	})

	go func() {
		common.Init(ui.state, ui.setVersions)

		time.Sleep(1 * time.Second)

		up, ver := update.CheckForUpdate(ui.state)
		if up {
			updatePopup(ver, ui)
		}
	}()

	ui.mainWindow.SetFixedSize(true)
	ui.resizeMainWindow()
	ui.mainWindow.CenterOnScreen()
	ui.mainWindow.ShowAndRun()
}
