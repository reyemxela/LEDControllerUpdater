package state

import (
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/reyemxela/LEDControllerUpdater/layout"
	"github.com/reyemxela/LEDControllerUpdater/releases"
	"github.com/sirupsen/logrus"
)

const (
	APP_NAME     = "LED Controller Updater"
	APP_VERSION  = "v1.2.0"
	TMP_DIR_NAME = "LEDControllerUpdater"
)

type State struct {
	Instance *rpc.Instance
	Ready    Ready
	TmpDir   string

	Versions       releases.Versions
	CurrentVersion string
	CurrentLayout  string

	CustomLayout   *layout.CustomLayout
	CustomSelected bool

	Ports       map[string]*rpc.Port
	CurrentPort string

	StatusFunc func(text string)

	AppType string
}

type Ready struct {
	PortSelected       bool
	NotFlashing        bool
	LibrariesInstalled bool
	CoreInstalled      bool
}

func NewState(appType string, statusFunc func(text string)) (*State, error) {
	s := &State{}

	s.AppType = appType
	s.StatusFunc = statusFunc

	// arduino-cli config
	configuration.Settings = configuration.Init("")
	logrus.SetLevel(logrus.FatalLevel)
	s.Instance = instance.CreateAndInit()

	s.CustomLayout = layout.DefaultLayout()
	s.CustomSelected = false

	tmpDir := os.TempDir()
	if tmpDir != "" {
		tmpDir = filepath.Join(tmpDir, TMP_DIR_NAME)
		os.MkdirAll(tmpDir, 0777)
		os.Chdir(tmpDir)
	}
	s.TmpDir = tmpDir

	s.Ports = make(map[string]*rpc.Port)

	s.Ready = Ready{
		NotFlashing: true,
	}

	return s, nil
}

func (s *State) SetStatus(text string) {
	if s.StatusFunc != nil {
		s.StatusFunc(text)
	}
}

func (s *State) CheckReady() bool {
	switch {
	case !s.Ready.PortSelected:
		s.SetStatus("No port selected")
	case !s.Ready.CoreInstalled:
		s.SetStatus("Arduino core still installing")
	case !s.Ready.LibrariesInstalled:
		s.SetStatus("Arduino libraries still installing")
	case !s.Ready.NotFlashing:
	default:
		return true
	}
	return false
}
