package main

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	mainWinWidth = 400
)

type Ready struct {
	Port               bool
	SelectionExists    bool
	LibrariesInstalled bool
	CoreInstalled      bool
	NotFlashing        bool
}

func (a *App) CheckReady() {
	if a.ready.Port &&
		a.ready.SelectionExists &&
		a.ready.LibrariesInstalled &&
		a.ready.CoreInstalled &&
		a.ready.NotFlashing {
		a.flashBtn.Enable()
	} else {
		if !a.batchRunning {
			a.flashBtn.Disable()
		}
	}
}

func (a *App) ResizeMainWindow() {
	a.mainWindow.Resize(fyne.Size{
		Width: mainWinWidth,

		// re-fit the window based on the contents, plus some
		// padding (the size jumps around a bit if there's none)
		Height: a.mainWindow.Content().MinSize().Height + 10,
	})
}

func (a *App) MakeMainSection() *fyne.Container {
	a.verSelect = widget.NewSelect([]string{}, func(value string) {
		a.UpdateLayouts(value)
	})
	a.verSelect.PlaceHolder = "(Select a version)"

	a.layoutSelect = widget.NewSelect([]string{}, func(value string) {
		if value == "-Custom-" {
			a.customSection.Show()
			a.ResizeMainWindow()
			a.ready.SelectionExists = true
		} else {
			a.customSection.Hide()
			a.ResizeMainWindow()
			if _, ok := a.releases[a.verSelect.Selected][value]; ok {
				a.ready.SelectionExists = true
			} else {
				a.ready.SelectionExists = false
			}
		}
		a.CheckReady()
	})
	a.layoutSelect.PlaceHolder = "(Select a layout)"

	a.portSelect = widget.NewSelect([]string{}, func(value string) {
		a.port = a.allPorts[value]
	})
	a.portSelect.PlaceHolder = "(Select COM port)"

	a.flashBtn = widget.NewButton("Flash Firmware", func() {
		if a.batchMode {
			if a.batchRunning {
				a.flashBtn.SetText("Start Batch")
				a.batchRunning = false
				a.batchPort = nil
				return
			} else {
				a.flashBtn.SetText("Stop Batch")
				a.batchRunning = true
				a.batchPort = a.port
			}
		}
		a.DoFlash()
	})
	a.flashBtn.Disable()

	driverBtn := widget.NewButton("CH340 Drivers", func() {
		go a.InstallCH340()
	})
	// CH340 driver button on windows only
	if runtime.GOOS != "windows" {
		driverBtn.Hide()
	}

	titleLabel := widget.NewLabel("WingnutTech LED Controller Updater " + APP_VERSION)
	titleLabel.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		titleLabel,
		driverBtn,
		a.verSelect,
		a.layoutSelect,
		container.NewGridWithColumns(2,
			container.NewVBox(
				a.portSelect,
				widget.NewLabel(""),
			),
			a.flashBtn,
		),
	)
}

func (a *App) MakeCustomSection() *fyne.Container {
	wingLEDLabel := widget.NewLabel("Wing: ")
	wingRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.WingRev = checked
	})
	wingLEDSlider := widget.NewSlider(0, 50)
	wingLEDSlider.OnChanged = func(value float64) {
		wingLEDLabel.SetText("Wing: " + fmt.Sprint(value))
		a.layout.WingLEDs = int(value)
	}
	wingLEDSlider.SetValue(19)

	noseLEDLabel := widget.NewLabel("Nose: ")
	noseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.NoseRev = checked
	})
	noseLEDSlider := widget.NewSlider(0, 50)
	noseLEDSlider.OnChanged = func(value float64) {
		noseLEDLabel.SetText("Nose: " + fmt.Sprint(value))
		a.layout.NoseLEDs = int(value)
	}
	noseRevCheck.SetChecked(true)
	noseLEDSlider.SetValue(6)

	fuseLEDLabel := widget.NewLabel("Fuse: ")
	fuseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.FuseRev = checked
	})
	fuseLEDSlider := widget.NewSlider(0, 50)
	fuseLEDSlider.OnChanged = func(value float64) {
		fuseLEDLabel.SetText("Fuse: " + fmt.Sprint(value))
		a.layout.FuseLEDs = int(value)
	}
	fuseLEDSlider.SetValue(13)

	tailLEDLabel := widget.NewLabel("Tail: ")
	tailRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.TailRev = checked
	})
	tailLEDSlider := widget.NewSlider(0, 50)
	tailLEDSlider.OnChanged = func(value float64) {
		tailLEDLabel.SetText("Tail: " + fmt.Sprint(value))
		a.layout.TailLEDs = int(value)
	}
	tailLEDSlider.SetValue(4)

	NavLEDLabel := widget.NewLabel("Nav LEDs: ")
	NavLEDSlider := widget.NewSlider(0, 50)
	NavLEDSlider.OnChanged = func(value float64) {
		NavLEDLabel.SetText("Nav LEDs: " + fmt.Sprint(value))
		a.layout.WingNavLEDs = int(value)
	}
	NavLEDSlider.SetValue(7)

	return container.NewVBox(
		container.NewGridWithColumns(2, wingLEDLabel, wingRevCheck),
		wingLEDSlider,
		container.NewGridWithColumns(2, noseLEDLabel, noseRevCheck),
		noseLEDSlider,
		container.NewGridWithColumns(2, fuseLEDLabel, fuseRevCheck),
		fuseLEDSlider,
		container.NewGridWithColumns(2, tailLEDLabel, tailRevCheck),
		tailLEDSlider,

		NavLEDLabel,
		NavLEDSlider,
	)
}

func (a *App) ToggleBatchMode() {
	if a.batchMode {
		a.batchMode = false
		a.flashBtn.SetText("Flash Firmware")
	} else {
		a.batchMode = true
		a.flashBtn.SetText("Start Batch")
	}
	a.batchPort = nil
	a.batchRunning = false
}

func (a *App) SetStatus(s string) {
	a.status.SetText(s)
}

func (a *App) NewPopup(t string, s string) {
	popup := a.fyneApp.NewWindow(t)
	label := widget.NewLabel(s)
	label.Wrapping = fyne.TextWrapWord

	popup.SetContent(
		container.NewVBox(
			label,
			layout.NewSpacer(),
			widget.NewButton("OK", func() {
				popup.Close()
			}),
		),
	)
	popup.Resize(fyne.NewSize(300, 100))
	popup.SetFixedSize(true)
	popup.CenterOnScreen()
	popup.Show()
}
