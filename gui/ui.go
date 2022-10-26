package main

import (
	"fmt"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/reyemxela/LEDControllerUpdater/arduino"
	"github.com/reyemxela/LEDControllerUpdater/common"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/update"
	"github.com/reyemxela/LEDControllerUpdater/utils"
)

type UI struct {
	app   fyne.App
	state *state.State

	verSelect    *widget.Select
	layoutSelect *widget.Select

	mainWindow    fyne.Window
	customSection *fyne.Container
	flashSection  *fyne.Container

	portList *widget.Select

	statusBar *widget.Label
}

func createVerSelect(ui *UI) {
	ui.verSelect = widget.NewSelect(nil, func(value string) {
		ui.state.CurrentVersion = value
		ui.layoutSelect.Options = append(utils.ListKeys(ui.state.Versions[value]), "-Custom-")
		ui.layoutSelect.SetSelectedIndex(0)
	})
}

func createLayoutSelect(ui *UI) {
	ui.layoutSelect = widget.NewSelect([]string{}, func(value string) {
		ui.state.CurrentLayout = value
		if value == "-Custom-" {
			ui.showCustomSection()
		} else {
			ui.hideCustomSection()
		}
	})
}

func createFlashSection(ui *UI) {
	ui.portList = widget.NewSelect([]string{}, func(value string) {
		ui.state.CurrentPort = value
		if value == "" {
			ui.state.Ready.PortSelected = false
		} else {
			ui.state.Ready.PortSelected = true
		}
	})
	ui.portList.PlaceHolder = "(Select COM port)"

	flashBtn := widget.NewButton("Flash Firmware", func() {
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

	ui.flashSection = container.NewGridWithColumns(2,
		container.NewVBox(
			ui.portList,
			widget.NewLabel(""),
		),
		flashBtn,
	)
}

func createCustomSection(ui *UI) {
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
		ui.state.CustomLayout.WingRev = checked
	})
	noseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		ui.state.CustomLayout.NoseRev = checked
	})
	fuseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		ui.state.CustomLayout.FuseRev = checked
	})
	tailRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		ui.state.CustomLayout.TailRev = checked
	})
	noseFuseJoinCheck := widget.NewCheck("Nose/Fuse joined?", func(checked bool) {
		ui.state.CustomLayout.NoseFuseJoin = checked
	})

	wingLEDSlider.OnChanged = func(value float64) {
		wingLEDLabel.SetText("Wing: " + fmt.Sprint(value))
		ui.state.CustomLayout.WingLEDs = int(value)
		navLEDSlider.Max = value
		if navLEDSlider.Value > value {
			navLEDSlider.OnChanged(value)
		}
		navLEDSlider.Refresh()
	}

	noseLEDSlider.OnChanged = func(value float64) {
		noseLEDLabel.SetText("Nose: " + fmt.Sprint(value))
		ui.state.CustomLayout.NoseLEDs = int(value)
	}

	fuseLEDSlider.OnChanged = func(value float64) {
		fuseLEDLabel.SetText("Fuse: " + fmt.Sprint(value))
		ui.state.CustomLayout.FuseLEDs = int(value)
	}

	tailLEDSlider.OnChanged = func(value float64) {
		tailLEDLabel.SetText("Tail: " + fmt.Sprint(value))
		ui.state.CustomLayout.TailLEDs = int(value)
	}

	navLEDSlider.OnChanged = func(value float64) {
		navLEDLabel.SetText("Nav LEDs: " + fmt.Sprint(value))
		ui.state.CustomLayout.WingNavLEDs = int(value)
	}

	wingLEDSlider.SetValue(float64(ui.state.CustomLayout.WingLEDs))
	noseLEDSlider.SetValue(float64(ui.state.CustomLayout.NoseLEDs))
	fuseLEDSlider.SetValue(float64(ui.state.CustomLayout.FuseLEDs))
	tailLEDSlider.SetValue(float64(ui.state.CustomLayout.TailLEDs))
	navLEDSlider.SetValue(float64(ui.state.CustomLayout.WingNavLEDs))
	noseRevCheck.SetChecked(true)
	noseFuseJoinCheck.SetChecked(true)

	label := widget.NewLabel("Custom settings:")
	label.Alignment = fyne.TextAlignCenter

	ui.customSection = container.NewHBox(
		widget.NewSeparator(),
		container.NewVBox(
			label,
			container.NewGridWithColumns(3, wingLEDLabel, layout.NewSpacer(), wingRevCheck),
			wingLEDSlider,
			widget.NewSeparator(),
			container.NewGridWithColumns(3, noseLEDLabel, layout.NewSpacer(), noseRevCheck),
			noseLEDSlider,
			widget.NewSeparator(),
			container.NewGridWithColumns(3, fuseLEDLabel, layout.NewSpacer(), fuseRevCheck),
			fuseLEDSlider,
			widget.NewSeparator(),
			container.NewGridWithColumns(3, tailLEDLabel, layout.NewSpacer(), tailRevCheck),
			tailLEDSlider,
			widget.NewSeparator(),

			navLEDLabel,
			navLEDSlider,
			widget.NewSeparator(),

			noseFuseJoinCheck,
		),
	)
	ui.customSection.Hide()
}

func createMainWindow(ui *UI) *fyne.Container {
	createVerSelect(ui)
	createLayoutSelect(ui)
	createFlashSection(ui)
	createCustomSection(ui)

	titleLabel := widget.NewLabel("WingnutTech LED Controller Updater " + state.APP_VERSION)
	titleLabel.Alignment = fyne.TextAlignCenter

	driverBtn := widget.NewButton("CH340 Drivers", func() {
		go common.InstallCH340(ui.state)
	})
	// CH340 driver button on windows only
	if runtime.GOOS != "windows" {
		driverBtn.Hide()
	}

	ui.statusBar = widget.NewLabel("")

	mainSection := container.NewVBox(
		titleLabel,
		driverBtn,
		ui.verSelect,
		ui.layoutSelect,
		ui.flashSection,
	)

	mainPlusCustom := container.NewHBox(
		mainSection,
		ui.customSection,
	)

	return container.NewVBox(mainPlusCustom, ui.statusBar)
}

func (ui *UI) setVersions() {
	ui.verSelect.Options = append(ui.verSelect.Options, utils.ListKeys(ui.state.Versions)...)
	ui.verSelect.SetSelectedIndex(0)
}

func updatePopup(ver string, ui *UI) {
	popup := ui.app.NewWindow("Update")
	var content *fyne.Container

	if runtime.GOOS == "darwin" {
		label := widget.NewLabel("New version available: " + ver + ".\n")
		link, _ := url.Parse(fmt.Sprintf("%s/%s/%s", update.APP_RELEASE_URL_PREFIX, "tag", ver))
		hyperlink := widget.NewHyperlink("Download", link)

		content = container.NewVBox(
			label,
			container.NewHBox(
				hyperlink,
				layout.NewSpacer(),
			),
			widget.NewButton("OK", func() {
				popup.Close()
			}),
		)
	} else {
		label := widget.NewLabel("New version available: " + ver + ".\n\nWould you like to automatically install the update?")
		content = container.NewVBox(
			label,
			layout.NewSpacer(),
			container.NewGridWithColumns(2,
				widget.NewButton("No", func() {
					popup.Close()
				}),
				widget.NewButton("Yes", func() {
					popup.Close()

					err := update.UpdateApp(ver, ui.state, func() { ui.app.Quit() })
					if err != nil {
						ui.state.SetStatus(err.Error())
					}
				}),
			),
		)
	}

	popup.SetContent(content)
	popup.CenterOnScreen()
	popup.Show()
}

func (ui *UI) setStatus(text string) {
	ui.statusBar.SetText(text)
}

func (ui *UI) resizeMainWindow() {
	ui.mainWindow.Resize(ui.mainWindow.Content().MinSize())
}

func (ui *UI) showCustomSection() {
	ui.customSection.Show()
	ui.state.CustomSelected = true
	ui.resizeMainWindow()
}

func (ui *UI) hideCustomSection() {
	ui.customSection.Hide()
	ui.state.CustomSelected = false
	ui.resizeMainWindow()
}
