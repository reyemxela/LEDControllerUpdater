package main

import (
	"fmt"
	"time"

	"github.com/reyemxela/LEDControllerUpdater/arduino"
	"github.com/reyemxela/LEDControllerUpdater/common"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/update"
	"github.com/reyemxela/LEDControllerUpdater/utils"
	"github.com/rivo/tview"
)

func main() {
	ui := &UI{}
	s, err := state.NewState("CLI", ui.setStatus)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ui.state = s

	ui.app = tview.NewApplication().EnableMouse(true)

	mainWindow := createMainWindow(ui)

	go arduino.WatchPorts(ui.state, func() {
		ui.app.QueueUpdateDraw(func() {
			ui.clearPortList()

			if len(ui.state.Ports) < 1 {
				ui.portList.AddOption(" -No Ports- ", nil)
				ui.portList.SetCurrentOption(0)
				return
			}

			portNames := utils.ListKeys(ui.state.Ports)
			selected := 0
			for i, p := range portNames {
				ui.portList.AddOption(p, nil)
				if p == ui.state.CurrentPort {
					selected = i
				}
			}
			ui.portList.SetCurrentOption(selected)
		})
	})

	go func() {
		common.Init(ui.state, ui.setVersions)

		time.Sleep(1 * time.Second)

		up, ver := update.CheckForUpdate(ui.state)
		if up {
			ui.app.QueueUpdateDraw(func() {
				ui.verSelect.InsertItem(-2, UPDATE_TEXT, "", 0, func() {
					err := update.UpdateApp(ver, ui.state, func() { ui.app.Stop() })
					if err != nil {
						s.SetStatus(err.Error())
					}
				})
			})
		}
	}()

	if err := ui.app.SetRoot(mainWindow, true).Run(); err != nil {
		panic(err)
	}
}
