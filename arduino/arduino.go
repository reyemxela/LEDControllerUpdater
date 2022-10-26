package arduino

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/reyemxela/LEDControllerUpdater/layout"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/utils"
	"go.bug.st/serial"
)

const (
	FQBN           = "arduino:avr:nano:cpu=atmega328"
	FQBNold        = "arduino:avr:nano:cpu=atmega328old"
	ZIP_URL_PREFIX = "https://github.com/wingnut-tech/LEDController/archive/refs/tags/"
)

var neededLibraries = [][]string{
	{"FastLED", "3.4.0"},
	{"Adafruit BMP280 Library", "2.3.0"},
}

func CheckLibraries(instance *rpc.Instance) error {
	for _, libs := range neededLibraries {
		if err := lib.LibraryInstall(context.Background(), &rpc.LibraryInstallRequest{
			Instance: instance,
			Name:     libs[0],
			Version:  libs[1],
		}, output.NewNullDownloadProgressCB(), output.NewNullTaskProgressCB()); err != nil {
			return err
		}
	}
	return nil
}

func CheckCore(instance *rpc.Instance) error {
	if _, err := core.PlatformInstall(context.Background(), &rpc.PlatformInstallRequest{
		Instance:        instance,
		PlatformPackage: "arduino",
		Architecture:    "avr",
	}, output.NewNullDownloadProgressCB(), output.NewNullTaskProgressCB()); err != nil {
		return err
	}
	return nil
}

func WatchPorts(s *state.State, callback func()) {
	eventsChan, _, err := board.Watch(&rpc.BoardListWatchRequest{Instance: s.Instance})
	if err != nil {
		s.SetStatus(err.Error())
	}

	// loop forever listening for board.Watch to give us events
	for event := range eventsChan {
		port := event.Port.Port
		addr := port.Address
		if event.EventType == "add" {
			s.Ports[addr] = port
			s.CurrentPort = addr
		} else {
			// if the port was in the list, remove it
			if _, ok := s.Ports[addr]; ok {
				delete(s.Ports, addr)
			}
			// if the unplugged board was our current one, grab the first available (if there is one)
			if s.CurrentPort == addr {
				if len(s.Ports) > 0 {
					for p := range s.Ports {
						s.CurrentPort = p
						break
					}
				} else {
					s.CurrentPort = ""
				}
			}
		}
		callback()
	}
}

func DoFlash(s *state.State) error {
	s.Ready.NotFlashing = false
	defer func() {
		s.Ready.NotFlashing = true
	}()

	if s.CustomSelected {
		return CompileAndFlash(s)
	} else {
		return DownloadAndFlash(s)
	}
}

func CompileAndFlash(s *state.State) error {
	ver := s.CurrentVersion

	newFolder := filepath.Join(s.TmpDir, ver)
	layoutFile := filepath.Join(newFolder, "layout.h")
	exportDir := filepath.Join(newFolder, "build")

	// if the ver folder doesn't already exist, download and unzip
	if _, err := os.Stat(newFolder); err != nil {
		s.SetStatus("Downloading " + ver)
		zipFile := ver + ".zip"
		zipUrl := ZIP_URL_PREFIX + zipFile

		err := utils.DownloadFile(zipFile, zipUrl)
		if err != nil {
			return err
		}

		fileNames, err := utils.UnzipFile(zipFile, s.TmpDir)
		if err != nil {
			return err
		}

		// assume first element is parent directory
		os.Rename(fileNames[0], newFolder)

		// rename .ino file because the arduino tools demand it matches the folder name
		os.Rename(
			filepath.Join(newFolder, "LEDController.ino"),
			filepath.Join(newFolder, ver+".ino"),
		)
	}
	s.SetStatus("Compiling custom " + ver + " layout...")

	// write out the custom layout into layout.h
	err := os.WriteFile(layoutFile, layout.GenerateCustomLayout(s.CustomLayout), os.ModePerm)
	if err != nil {
		return err
	}

	if _, err := compile.Compile(context.Background(), &rpc.CompileRequest{
		Instance:   s.Instance,
		Fqbn:       FQBN,
		SketchPath: newFolder,
		ExportDir:  exportDir,
	}, io.Discard, io.Discard, nil, false); err != nil {
		return err
	}

	s.SetStatus("Flashing custom " + ver + " layout...")
	if err := FlashHex(filepath.Join(exportDir, ver+".ino.hex"), s.Instance, s.Ports[s.CurrentPort]); err != nil {
	}

	return nil
}

func DownloadAndFlash(s *state.State) error {
	ver, lay := s.CurrentVersion, s.CurrentLayout
	hexFile := filepath.Join(s.TmpDir, lay)
	hexURL := s.Versions[ver][lay]
	if _, err := os.Stat(hexFile); err != nil {
		s.SetStatus("Downloading " + lay)
		if err := utils.DownloadFile(hexFile, hexURL); err != nil {
			return err
		}
	}
	s.SetStatus("Flashing " + lay + "...")
	if err := FlashHex(hexFile, s.Instance, s.Ports[s.CurrentPort]); err != nil {
		return err
	}
	return nil
}

func FlashHex(hexFile string, instance *rpc.Instance, port *rpc.Port) error {
	bl := FQBNold

	tb, err := testBootloaderType(port.Address, 115200)
	if err != nil {
		return err
	}
	if tb {
		bl = FQBN
	} else {
		tb, err := testBootloaderType(port.Address, 57600)
		if err != nil {
			return err
		}
		if tb {
		} else {
		}
	}

	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   instance,
		Fqbn:       bl,
		SketchPath: filepath.Dir(hexFile),
		Port:       port,
		ImportFile: hexFile,
	}, io.Discard, io.Discard); err != nil {
		return err
	}

	return nil
}

func testBootloaderType(p string, b int) (bool, error) {
	syncCmd := []byte{0x30, 0x20}
	inSyncResp := []byte{0x14, 0x10}
	delay := (250 * time.Millisecond)
	shortDelay := (50 * time.Millisecond)
	timeout := (250 * time.Millisecond)

	port, err := serial.Open(p, &serial.Mode{BaudRate: b})
	if err != nil {
		return false, err
	}
	defer port.Close()

	port.SetReadTimeout(timeout)

	// reset bootloader
	port.SetDTR(false)
	port.SetRTS(false)
	time.Sleep(delay)

	port.SetDTR(true)
	port.SetRTS(true)
	time.Sleep(shortDelay)

	port.ResetInputBuffer()

	for i := 0; i < 4; i++ {
		port.Write(syncCmd)
		time.Sleep(shortDelay)

		resp := make([]byte, 2)
		port.Read(resp)
		if bytes.Equal(resp, inSyncResp) {
			return true, nil
		}
		port.ResetInputBuffer()
	}
	return false, nil
}
