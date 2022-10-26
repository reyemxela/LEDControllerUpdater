package main

import (
	"runtime"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/reyemxela/LEDControllerUpdater/arduino"
	"github.com/reyemxela/LEDControllerUpdater/common"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/utils"
	"github.com/rivo/tview"
)

const (
	QUIT_TEXT   = "quit"
	UPDATE_TEXT = "Update Available!"
	CH340_TEXT  = "CH340 drivers"
	SEPARATOR   = "------"
)

type UI struct {
	app   *tview.Application
	state *state.State

	verSelect    *tview.List
	layoutSelect *tview.List

	ledForm      *tview.Form
	checkboxForm *tview.Form

	customSection *tview.Pages
	flashSection  *tview.Flex

	flashButton *tview.Button
	portList    *tview.DropDown

	statusBar *tview.TextView

	customEnabled bool

	flowWithCustom    []tview.Primitive
	flowWithoutCustom []tview.Primitive
}

func (ui *UI) setStatus(text string) {
	ui.statusBar.SetText(text)
}

func createFlows(ui *UI) {
	ui.flowWithCustom = []tview.Primitive{
		ui.verSelect,
		ui.layoutSelect,
		ui.ledForm,
		ui.checkboxForm,
		ui.portList,
		ui.flashButton,
	}

	ui.flowWithoutCustom = []tview.Primitive{
		ui.verSelect,
		ui.layoutSelect,
		ui.portList,
		ui.flashButton,
	}
}

func createVerSelect(ui *UI) {
	ui.verSelect = tview.NewList().ShowSecondaryText(false)
	ui.verSelect.SetBorder(true).SetTitle("Version")
	ui.verSelect.SetChangedFunc(func(i int, text, _ string, _ rune) {
		ui.layoutSelect.Clear()
		if text == QUIT_TEXT || text == UPDATE_TEXT || text == CH340_TEXT || text == SEPARATOR {
			return
		}

		for _, l := range utils.ListKeys(ui.state.Versions[text]) {
			ui.layoutSelect.AddItem(l, "", 0, nil)
		}
		ui.layoutSelect.AddItem("-Custom-", "", 0, nil)
		ui.state.CurrentVersion = text
	})
}

func createLayoutSelect(ui *UI) {
	ui.layoutSelect = tview.NewList().ShowSecondaryText(false)
	ui.layoutSelect.SetBorder(true).SetTitle("Layout")

	ui.layoutSelect.SetChangedFunc(func(i int, text, _ string, _ rune) {
		ui.state.CurrentLayout = text
		if text == "-Custom-" {
			ui.customSection.SwitchToPage("Custom")
			ui.customEnabled = true
			ui.state.CustomSelected = true
		} else {
			ui.customSection.SwitchToPage("Blank")
			ui.customEnabled = false
			ui.state.CustomSelected = false
		}
	})
}

func createLedForm(ui *UI) {
	layoutValidate := func(textToCheck string, lastChar rune) bool {
		if len(textToCheck) > 2 {
			return false
		}
		if _, err := strconv.Atoi(textToCheck); err != nil {
			return false
		}
		return true
	}

	ui.ledForm = tview.NewForm().
		AddInputField("Wing LEDs:", strconv.Itoa(ui.state.CustomLayout.WingLEDs), 4, layoutValidate, func(text string) {
			ui.state.CustomLayout.WingLEDs, _ = strconv.Atoi("0" + text)
		}).
		AddInputField("Nose LEDs:", strconv.Itoa(ui.state.CustomLayout.NoseLEDs), 4, layoutValidate, func(text string) {
			ui.state.CustomLayout.NoseLEDs, _ = strconv.Atoi("0" + text)
		}).
		AddInputField("Fuse LEDs:", strconv.Itoa(ui.state.CustomLayout.FuseLEDs), 4, layoutValidate, func(text string) {
			ui.state.CustomLayout.FuseLEDs, _ = strconv.Atoi("0" + text)
		}).
		AddInputField("Tail LEDs:", strconv.Itoa(ui.state.CustomLayout.TailLEDs), 4, layoutValidate, func(text string) {
			ui.state.CustomLayout.TailLEDs, _ = strconv.Atoi("0" + text)
		}).
		AddInputField("Nav LEDs:", strconv.Itoa(ui.state.CustomLayout.WingNavLEDs), 4, layoutValidate, func(text string) {
			ui.state.CustomLayout.WingNavLEDs, _ = strconv.Atoi("0" + text)
		})
}

func createCheckboxForm(ui *UI) {
	ui.checkboxForm = tview.NewForm().
		AddCheckbox("Reverse:", ui.state.CustomLayout.WingRev, func(checked bool) {
			ui.state.CustomLayout.WingRev = checked
		}).
		AddCheckbox("Reverse:", ui.state.CustomLayout.NoseRev, func(checked bool) {
			ui.state.CustomLayout.NoseRev = checked
		}).
		AddCheckbox("Reverse:", ui.state.CustomLayout.FuseRev, func(checked bool) {
			ui.state.CustomLayout.FuseRev = checked
		}).
		AddCheckbox("Reverse:", ui.state.CustomLayout.TailRev, func(checked bool) {
			ui.state.CustomLayout.TailRev = checked
		}).
		AddCheckbox("Nose/Fuse join:", ui.state.CustomLayout.NoseFuseJoin, func(checked bool) {
			ui.state.CustomLayout.NoseFuseJoin = checked
		})
}

func createFlashSection(ui *UI) {
	ui.portList = tview.NewDropDown().
		AddOption(" -No Ports- ", nil).
		SetCurrentOption(0).
		SetLabel("Port: ").SetTextOptions("", "", "", "", " -None-")
	ui.portList.SetSelectedFunc(func(text string, index int) {
		ui.state.CurrentPort = text
		if text == " -No Ports- " {
			ui.state.Ready.PortSelected = false
		} else {
			ui.state.Ready.PortSelected = true
		}
	})

	ui.flashButton = tview.NewButton("Flash")
	ui.flashButton.SetSelectedFunc(func() {
		if ui.state.CheckReady() {
			go func() {
				ui.state.Ready.NotFlashing = false
				err := arduino.DoFlash(ui.state)
				if err != nil {
					ui.state.SetStatus(err.Error())
				} else {
					ui.state.SetStatus("Done!")
				}
				ui.state.Ready.NotFlashing = true
			}()
		}
	})

	ui.flashSection = tview.NewFlex().
		AddItem(ui.portList, 0, 2, false).
		AddItem(ui.flashButton, 9, 0, false)
	ui.flashSection.SetBorder(true)
}

func createCustomSection(ui *UI) {
	ui.customSection = tview.NewPages().
		AddPage("Custom", tview.NewFlex().
			AddItem(ui.ledForm, 0, 1, false).
			AddItem(ui.checkboxForm, 0, 1, false), true, false).
		AddPage("Blank", tview.NewBox(), true, true)
	ui.customSection.SetBorder(true).SetTitle("Custom layout")

	ui.customEnabled = false
}

func createMainWindow(ui *UI) *tview.Flex {
	createVerSelect(ui)
	createLayoutSelect(ui)
	createLedForm(ui)
	createCheckboxForm(ui)
	createFlashSection(ui)
	createCustomSection(ui)

	titleBar := tview.NewTextView().SetText(state.APP_NAME + " " + state.APP_VERSION).SetTextAlign(tview.AlignCenter)
	ui.statusBar = tview.NewTextView().SetText("")
	ui.statusBar.SetChangedFunc(func() { ui.app.Draw() })

	mainWindow := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleBar, 2, 0, false).
		AddItem(
			tview.NewFlex().
				AddItem(ui.verSelect, 0, 1, true).
				AddItem(ui.layoutSelect, 0, 1, false).
				AddItem(
					tview.NewFlex().
						AddItem(ui.customSection, 0, 1, false).
						AddItem(ui.flashSection, 5, 0, false).
						SetDirection(tview.FlexRow),
					0, 2, false),
			0, 1, true).
		AddItem(ui.statusBar, 1, 0, false)

	setupInputs(ui)
	createFlows(ui)

	return mainWindow
}

func setupInputs(ui *UI) {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight {
			ui.move(1)
			return nil
		} else if event.Key() == tcell.KeyLeft {
			ui.move(-1)
			return nil
		} else if ui.ledForm.HasFocus() || ui.checkboxForm.HasFocus() {
			if event.Key() == tcell.KeyDown {
				return tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone)
			} else if event.Key() == tcell.KeyUp {
				return tcell.NewEventKey(tcell.KeyBacktab, ' ', tcell.ModNone)
			}
		}
		return event
	})
}

func (ui *UI) move(dir int) {
	var flow []tview.Primitive
	if ui.customEnabled {
		flow = ui.flowWithCustom
	} else {
		flow = ui.flowWithoutCustom
	}
	l := len(flow)
	if ui.app.GetFocus() == nil {
		ui.app.SetFocus(flow[0])
		return
	}

	for i := range flow {
		if flow[i].HasFocus() {
			i = ((i+dir)%l + l) % l

			ui.app.SetFocus(flow[i])
			return
		}
	}
	ui.app.SetFocus(flow[0])
}

func (ui *UI) clearPortList() {
	for ui.portList.GetOptionCount() > 0 {
		ui.portList.RemoveOption(0)
	}
}

func (ui *UI) setVersions() {
	ui.verSelect.Clear()
	for _, v := range utils.ListKeys(ui.state.Versions) {
		ui.verSelect.AddItem(v, "", 0, nil)
	}
	ui.verSelect.AddItem(SEPARATOR, "", 0, nil)
	if runtime.GOOS == "windows" {
		ui.verSelect.AddItem(CH340_TEXT, "", 0, func() {
			go common.InstallCH340(ui.state)
		})
	}
	ui.verSelect.AddItem("quit", "", 'q', func() {
		ui.app.Stop()
	})
}
