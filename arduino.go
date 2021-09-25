package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

const (
	FQBN    = "arduino:avr:nano:cpu=atmega328"
	FQBNold = "arduino:avr:nano:cpu=atmega328old"
)

var neededLibraries = [][]string{
	{"FastLED", "3.4.0"},
	{"Adafruit BMP280 Library", "2.3.0"},
}

func (a *App) CheckLibraries() {
	r := true
	for _, libs := range neededLibraries {
		if err := lib.LibraryInstall(context.Background(), &rpc.LibraryInstallRequest{
			Instance: a.instance,
			Name:     libs[0],
			Version:  libs[1],
		}, output.NewNullDownloadProgressCB(), output.NewNullTaskProgressCB()); err != nil {
			a.NewPopup("Error", err.Error())
			r = false
		}
	}
	a.ready.LibrariesInstalled = r
	a.CheckReady()
}

func (a *App) CheckCore() {
	if _, err := core.PlatformInstall(context.Background(), &rpc.PlatformInstallRequest{
		Instance:        a.instance,
		PlatformPackage: "arduino",
		Architecture:    "avr",
	}, output.NewNullDownloadProgressCB(), output.NewNullTaskProgressCB()); err != nil {
		a.NewPopup("Error", err.Error())
		return
	}
	a.ready.CoreInstalled = true
	a.CheckReady()
}

func (a *App) GetPorts() {
	eventsChan, err := board.Watch(a.instance.Id, nil)
	if err != nil {
		a.NewPopup("Error", err.Error())
	}

	// loop forever listening for board.Watch to give us events
	for event := range eventsChan {
		port := event.Port.Port
		addr := port.Address
		if event.EventType == "add" {
			// a board got plugged in. add it to the list, and set the current port to it
			a.allPorts[addr] = port
			a.portSelect.Options = append(a.portSelect.Options, addr)
			a.port = port

			if a.batchRunning && a.port.Address == a.batchPort.Address {
				a.DoFlash()
			}
		} else {
			// board got unplugged. remove it from the list
			for i, v := range a.portSelect.Options {
				if v == addr {
					a.portSelect.Options = append(a.portSelect.Options[:i], a.portSelect.Options[i+1:]...)
				}
			}
			delete(a.allPorts, addr)
			// if the unplugged board was our current one, grab the next available (if there is one)
			if a.port.Address == port.Address {
				if len(a.portSelect.Options) > 0 {
					a.port = a.allPorts[a.portSelect.Options[0]]
				} else {
					a.port = nil
				}
			}
		}
		if a.port == nil {
			a.ready.Port = false
			a.portSelect.ClearSelected()
		} else {
			a.ready.Port = true
			a.portSelect.SetSelected(addr)
		}
		a.CheckReady()
	}
}

func (a *App) DoFlash() {
	if a.layoutSelect.Selected == "-Custom-" {
		go a.CompileAndFlash(a.verSelect.Selected)
	} else {
		go a.DownloadAndFlash(a.verSelect.Selected, a.layoutSelect.Selected)
	}
}

func (a *App) CompileAndFlash(v string) {
	a.ready.NotFlashing = false
	a.CheckReady()

	defer func() {
		a.ready.NotFlashing = true
		a.CheckReady()
	}()

	newFolder := filepath.Join(a.tmpPath, v)
	layoutFile := filepath.Join(newFolder, "layout.h")
	exportDir := filepath.Join(newFolder, "build")

	if _, err := os.Stat(newFolder); err != nil {
		a.SetStatus("Downloading " + v)
		zipName := v + ".zip"
		zipFile := filepath.Join(a.tmpPath, zipName)
		zipUrl := ZIP_URL_PREFIX + zipName

		err := a.DownloadFile(zipFile, zipUrl)
		if err != nil {
			a.NewPopup("Error", err.Error())
			return
		}

		fileNames, err := a.UnzipFile(zipFile, a.tmpPath)
		if err != nil {
			a.NewPopup("Error", err.Error())
			return
		}

		// assume first element is parent directory
		os.Rename(fileNames[0], newFolder)

		// rename .ino file because the arduino tools demand it matches the folder name
		os.Rename(
			filepath.Join(newFolder, "LEDController.ino"),
			filepath.Join(newFolder, v+".ino"),
		)
	}

	a.SetStatus("Compiling custom firmware")

	// write out the custom layout into layout.h
	err := os.WriteFile(layoutFile, a.GenerateCustomLayout(), os.ModePerm)
	if err != nil {
		a.NewPopup("Error", err.Error())
		return
	}

	if _, err := compile.Compile(context.Background(), &rpc.CompileRequest{
		Instance:   a.instance,
		Fqbn:       FQBN,
		SketchPath: newFolder,
		ExportDir:  exportDir,
	}, io.Discard, io.Discard, false); err != nil {
		fmt.Println(err)
	}

	a.SetStatus("Flashing custom firmware")
	if err := a.FlashHex(filepath.Join(exportDir, v+".ino.hex")); err != nil {
		a.NewPopup("Error", err.Error())
	}

	a.ready.NotFlashing = true
	a.CheckReady()
}

func (a *App) DownloadAndFlash(v string, h string) {
	a.ready.NotFlashing = false
	a.CheckReady()

	defer func() {
		a.ready.NotFlashing = true
		a.CheckReady()
	}()

	hexFile := filepath.Join(a.tmpPath, h)
	if _, err := os.Stat(hexFile); err != nil {
		if err := a.DownloadFile(hexFile, a.releases[v][h]); err != nil {
			a.NewPopup("Error", err.Error())
			return
		}
	}
	if err := a.FlashHex(hexFile); err != nil {
		a.NewPopup("Error", err.Error())
	}
}

func (a *App) FlashHex(hexFile string) error {
	a.SetStatus("Flashing " + filepath.Base(hexFile))

	ul := func(f string) error {
		_, err := upload.Upload(context.Background(), &rpc.UploadRequest{
			Instance:   a.instance,
			Fqbn:       f,
			SketchPath: a.tmpPath,
			Port:       a.port,
			ImportFile: hexFile,
		}, io.Discard, io.Discard)
		return err
	}

	if err := ul(FQBN); err != nil {
		if strings.Contains(err.Error(), "uploading error") {
			a.SetStatus("Trying old bootloader mode")
			if err := ul(FQBNold); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	a.SetStatus("Done!")
	return nil
}
