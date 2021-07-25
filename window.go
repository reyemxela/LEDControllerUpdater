package main

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
		a.flashBtn.Disable()
	}
}

func (a *App) ResizeMainWindow() {
	a.mainWindow.Resize(fyne.Size{
		Width: mainWinWidth,

		// re-fit the window based on the contents, plus some
		// padding (the size jumps around a bit if there's none)
		Height: a.mainWindow.Content().MinSize().Height + 10,
	})

	a.mainWindow.CenterOnScreen()
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
		a.port = value
	})
	a.portSelect.PlaceHolder = "(Select COM port)"

	a.flashBtn = widget.NewButton("Flash Firmware", func() {
		if a.layoutSelect.Selected == "-Custom-" {
			go a.CompileAndFlash(a.verSelect.Selected)
		} else {
			go a.DownloadAndFlash(a.verSelect.Selected, a.layoutSelect.Selected)
		}
	})
	a.flashBtn.Disable()

	driverBtn := widget.NewButton("CH340 Drivers", func() {
		go a.InstallCH340()
	})
	// CH340 driver button on windows only
	if runtime.GOOS != "windows" {
		driverBtn.Hide()
	}

	return container.NewVBox(
		widget.NewLabel("WingnutTech LED Controller Updater "+APP_VERSION),
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
		a.config.WingRev = checked
	})
	wingLEDSlider := widget.NewSlider(0, 50)
	wingLEDSlider.OnChanged = func(value float64) {
		wingLEDLabel.SetText("Wing: " + fmt.Sprint(value))
		a.config.WingLEDs = int(value)
	}
	wingLEDSlider.SetValue(19)

	noseLEDLabel := widget.NewLabel("Nose: ")
	noseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.config.NoseRev = checked
	})
	noseLEDSlider := widget.NewSlider(0, 50)
	noseLEDSlider.OnChanged = func(value float64) {
		noseLEDLabel.SetText("Nose: " + fmt.Sprint(value))
		a.config.NoseLEDs = int(value)
	}
	noseRevCheck.SetChecked(true)
	noseLEDSlider.SetValue(6)

	fuseLEDLabel := widget.NewLabel("Fuse: ")
	fuseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.config.FuseRev = checked
	})
	fuseLEDSlider := widget.NewSlider(0, 50)
	fuseLEDSlider.OnChanged = func(value float64) {
		fuseLEDLabel.SetText("Fuse: " + fmt.Sprint(value))
		a.config.FuseLEDs = int(value)
	}
	fuseLEDSlider.SetValue(13)

	tailLEDLabel := widget.NewLabel("Tail: ")
	tailRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		a.config.TailRev = checked
	})
	tailLEDSlider := widget.NewSlider(0, 50)
	tailLEDSlider.OnChanged = func(value float64) {
		tailLEDLabel.SetText("Tail: " + fmt.Sprint(value))
		a.config.TailLEDs = int(value)
	}
	tailLEDSlider.SetValue(4)

	NavLEDLabel := widget.NewLabel("Nav LEDs: ")
	NavLEDSlider := widget.NewSlider(0, 50)
	NavLEDSlider.OnChanged = func(value float64) {
		NavLEDLabel.SetText("Nav LEDs: " + fmt.Sprint(value))
		a.config.WingNavLEDs = int(value)
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
