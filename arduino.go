package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

func (a *App) CheckLibraries() {
	r := true
	for _, libName := range neededLibraries {
		if err := lib.LibraryInstall(context.Background(), &rpc.LibraryInstallRequest{
			Instance: a.instance,
			Name:     libName,
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
		addr := event.Port.Address
		if event.EventType == "add" {
			// a board got plugged in. add it to the list, and set the current port to it
			a.portSelect.Options = append(a.portSelect.Options, addr)
			a.port = addr
		} else {
			// board got unplugged. remove it from the list
			for i, v := range a.portSelect.Options {
				if v == addr {
					a.portSelect.Options = append(a.portSelect.Options[:i], a.portSelect.Options[i+1:]...)
				}
			}
			// if the unplugged board was our current one, grab the next available (if there is one)
			if a.port == addr {
				if len(a.portSelect.Options) > 0 {
					a.port = a.portSelect.Options[0]
				} else {
					a.port = ""
				}
			}
		}
		if a.port != "" {
			a.ready.Port = true
			a.portSelect.SetSelected(a.port)
		} else {
			a.ready.Port = false
			a.portSelect.ClearSelected()
		}
		a.CheckReady()
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
	configFile := filepath.Join(newFolder, "config.h")
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
			filepath.Join(newFolder, "FT-Night-Radian-LED-Controller.ino"),
			filepath.Join(newFolder, v+".ino"),
		)
	}

	a.SetStatus("Compiling custom firmware")

	// write out the custom config into config.h
	err := os.WriteFile(configFile, a.GenerateCustomConfig(), os.ModePerm)
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
	err = a.FlashHex(filepath.Join(exportDir, v+".ino.hex"))
	if err != nil {
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
	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   a.instance,
		Fqbn:       FQBN,
		SketchPath: a.tmpPath,
		Port:       a.port,
		ImportFile: hexFile,
	}, io.Discard, io.Discard); err != nil {
		return err
	}

	a.SetStatus("Done!")
	return nil
}
