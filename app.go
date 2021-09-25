package main

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

type App struct {
	tmpPath  string
	port     *rpc.Port
	allPorts map[string]*rpc.Port
	releases map[string]map[string]string
	layout   CustomLayout
	ready    Ready

	batchMode    bool
	batchRunning bool
	batchPort    *rpc.Port

	instance *rpc.Instance

	fyneApp    fyne.App
	mainWindow fyne.Window

	verSelect    *widget.Select
	layoutSelect *widget.Select
	portSelect   *widget.Select

	flashBtn      *widget.Button
	status        *widget.Label
	customSection *fyne.Container
}

func CreateApp() *App {
	a := &App{}

	// arduino-cli config
	configuration.Settings = configuration.Init("")
	logrus.SetLevel(logrus.FatalLevel)
	a.instance = instance.CreateAndInit()

	a.tmpPath = os.TempDir()
	if a.tmpPath != "" {
		a.tmpPath = filepath.Join(a.tmpPath, TMP_DIR_NAME)
		os.MkdirAll(a.tmpPath, 0777)
	}

	a.releases = make(map[string]map[string]string)
	a.allPorts = make(map[string]*rpc.Port)
	a.layout = CustomLayout{}

	a.ready = Ready{
		NotFlashing: true,
	}

	a.fyneApp = app.New()
	a.mainWindow = a.fyneApp.NewWindow(APP_NAME)
	a.customSection = a.MakeCustomSection()
	a.status = widget.NewLabel("")

	a.mainWindow.SetContent(container.NewVBox(
		a.MakeMainSection(),
		a.customSection,
		a.status,
	))

	ctrlB := desktop.CustomShortcut{KeyName: fyne.KeyB, Modifier: desktop.ControlModifier | desktop.ShiftModifier}
	a.mainWindow.Canvas().AddShortcut(&ctrlB, func(shortcut fyne.Shortcut) {
		a.ToggleBatchMode()
	})

	return a
}
