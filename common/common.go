package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/reyemxela/LEDControllerUpdater/arduino"
	"github.com/reyemxela/LEDControllerUpdater/releases"
	"github.com/reyemxela/LEDControllerUpdater/state"
	"github.com/reyemxela/LEDControllerUpdater/utils"
)

const (
	CH340_URL = "https://github.com/reyemxela/LEDControllerUpdater/releases/download/v1.0.0/CH34x_Install_Windows_v3_4.zip"
	CH340_EXE = "CH34x_Install_Windows_v3_4.EXE"
)

func InstallCH340(s *state.State) {
	// ch340 drivers are only needed on windows, because windows is so special.
	if runtime.GOOS != "windows" {
		return
	}

	s.SetStatus("Downloading CH340 drivers")

	zipFile := filepath.Join(s.TmpDir, "ch340.zip")
	exe := filepath.Join(s.TmpDir, CH340_EXE)

	if _, err := os.Stat(exe); err != nil {
		if _, err := os.Stat(zipFile); err != nil {
			err := utils.DownloadFile(zipFile, CH340_URL)
			if err != nil {
				s.SetStatus(err.Error())
				return
			}
		}

		_, err = utils.UnzipFile(zipFile, s.TmpDir)
		if err != nil {
			s.SetStatus(err.Error())
			return
		}
	}

	time.Sleep(time.Second * 2)

	cmd := exec.Command(exe)
	err := cmd.Start()
	if err != nil {
		s.SetStatus(err.Error())
		return
	}
	s.SetStatus("Started CH340 installer")
}

func Init(s *state.State, setVersions func()) {
	s.SetStatus("Downloading versions...")
	v, err := releases.GetVersions()
	if err != nil {
		s.SetStatus("Error: " + err.Error())
	}
	s.Versions = v
	if setVersions != nil {
		setVersions()
	}

	s.SetStatus("Checking arduino core...")
	err = arduino.CheckCore(s.Instance)
	if err != nil {
		s.SetStatus("Error: " + err.Error())
	} else {
		s.Ready.CoreInstalled = true
	}

	s.SetStatus("Checking arduino libraries...")
	err = arduino.CheckLibraries(s.Instance)
	if err != nil {
		s.SetStatus("Error: " + err.Error())
	} else {
		s.Ready.LibrariesInstalled = true
	}

	s.SetStatus("Ready")
}
