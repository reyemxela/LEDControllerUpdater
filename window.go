package main

import (
	"fmt"
	"runtime"
	"time"

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
	oldSize := a.mainWindow.Canvas().Size()
	newSize := fyne.Size{
		Width: mainWinWidth,

		// re-fit the window based on the contents, plus some
		// padding (the size jumps around a bit if there's none)
		Height: float32(int(a.mainWindow.Content().MinSize().Height) + 10),
	}

	a.mainWindow.Resize(newSize)

	// it seems to take a small amount of time after calling Resize()
	// before the canvas size actually updates, so this just waits until
	// the new canvas size is applied before calling CenterOnScreen()
	if newSize != oldSize {
		go func() {
			for i := 0; i < 100; i++ { // limit to 100 loops just in case something goes funky
				if a.mainWindow.Canvas().Size() != oldSize {
					a.mainWindow.CenterOnScreen()
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
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
	noseLEDLabel := widget.NewLabel("Nose: ")
	fuseLEDLabel := widget.NewLabel("Fuse: ")
	tailLEDLabel := widget.NewLabel("Tail: ")
	navLEDLabel := widget.NewLabel("Nav LEDs: ")

	wingLEDSlider := widget.NewSlider(0, 50)
	noseLEDSlider := widget.NewSlider(0, 50)
	fuseLEDSlider := widget.NewSlider(0, 50)
	tailLEDSlider := widget.NewSlider(0, 50)
	navLEDSlider := widget.NewSlider(0, 50)

	wingRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.WingRev = checked
	})
	noseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.NoseRev = checked
	})
	fuseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.FuseRev = checked
	})
	tailRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.layout.TailRev = checked
	})
	noseFuseJoinCheck := widget.NewCheck("Nose/Fuse joined?", func(checked bool) {
		a.layout.NoseFuseJoin = checked
	})

	wingLEDSlider.OnChanged = func(value float64) {
		wingLEDLabel.SetText("Wing: " + fmt.Sprint(value))
		a.layout.WingLEDs = int(value)
		navLEDSlider.Max = value
		if navLEDSlider.Value > value {
			navLEDSlider.OnChanged(value)
		}
		navLEDSlider.Refresh()
	}

	noseLEDSlider.OnChanged = func(value float64) {
		noseLEDLabel.SetText("Nose: " + fmt.Sprint(value))
		a.layout.NoseLEDs = int(value)
	}

	fuseLEDSlider.OnChanged = func(value float64) {
		fuseLEDLabel.SetText("Fuse: " + fmt.Sprint(value))
		a.layout.FuseLEDs = int(value)
	}

	tailLEDSlider.OnChanged = func(value float64) {
		tailLEDLabel.SetText("Tail: " + fmt.Sprint(value))
		a.layout.TailLEDs = int(value)
	}

	navLEDSlider.OnChanged = func(value float64) {
		navLEDLabel.SetText("Nav LEDs: " + fmt.Sprint(value))
		a.layout.WingNavLEDs = int(value)
	}

	wingLEDSlider.SetValue(31)
	noseLEDSlider.SetValue(4)
	fuseLEDSlider.SetValue(18)
	tailLEDSlider.SetValue(8)
	navLEDSlider.SetValue(8)
	noseRevCheck.SetChecked(true)
	noseFuseJoinCheck.SetChecked(true)

	return container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(2, wingLEDLabel, wingRevCheck),
		wingLEDSlider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, noseLEDLabel, noseRevCheck),
		noseLEDSlider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, fuseLEDLabel, fuseRevCheck),
		fuseLEDSlider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, tailLEDLabel, tailRevCheck),
		tailLEDSlider,
		widget.NewSeparator(),

		navLEDLabel,
		navLEDSlider,
		widget.NewSeparator(),

		noseFuseJoinCheck,
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
