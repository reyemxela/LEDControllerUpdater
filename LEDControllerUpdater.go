package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

const API_URL = "https://api.github.com/repos/wingnut-tech/FT-Night-Radian-LED-Controller/releases"
const CH340_URL = "https://github.com/reyemxela/LEDControllerUpdater/releases/download/v1.0.0/CH34x_Install_Windows_v3_4.zip"
const ZIP_URL_PREFIX = "https://github.com/wingnut-tech/FT-Night-Radian-LED-Controller/archive/refs/tags/"
const FQBN = "arduino:avr:nano"
const mainWinWidth = 400

var neededLibraries = []string{"FastLED", "Adafruit BMP280 Library"}

var (
	tmpPath      string
	verSelect    *widget.Select
	layoutSelect *widget.Select
	portSelect   *widget.Select
	flashBtn     *widget.Button
	statusLabel  *widget.Label
	releases     map[string]map[string]string
	vboxMain     *fyne.Container
	vboxCustom   *fyne.Container
	mainWindow   fyne.Window
	mainApp      fyne.App
	customConfig CustomConfig
	inst         *rpc.Instance
	port         string
	ready        Ready
)

type CustomConfig struct {
	WingLEDs int
	NoseLEDs int
	FuseLEDs int
	TailLEDs int

	WingNavLEDs int

	WingRev bool
	NoseRev bool
	FuseRev bool
	TailRev bool
}

type Ready struct {
	Port               bool
	SelectionExists    bool
	LibrariesInstalled bool
	NotFlashing        bool
}

type Releases []struct {
	Name   string `json:"name"`
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {
	releases = make(map[string]map[string]string)

	tmpPath = os.TempDir()
	if tmpPath != "" {
		tmpPath = filepath.Join(tmpPath, "LEDController")
		os.MkdirAll(tmpPath, 0777)
	}

	// config and instance
	configuration.Settings = configuration.Init("")
	logrus.SetLevel(logrus.FatalLevel)
	inst = instance.CreateAndInit()

	ready.NotFlashing = true

	mainApp = app.New()
	mainWindow = mainApp.NewWindow("LED Controller Updater")

	vboxMain = makeMainSection()
	vboxCustom = makeCustomSection()
	vboxCustom.Hide()

	statusLabel = widget.NewLabel("")

	go updateReleases()
	go getPorts()
	go checkLibraries()

	mainWindow.SetContent(container.NewVBox(
		vboxMain,
		vboxCustom,
		statusLabel,
	))
	mainWindow.SetFixedSize(true)
	resizeMainWindow()
	mainWindow.CenterOnScreen()

	mainWindow.ShowAndRun()
}

func makeMainSection() *fyne.Container {
	topLabel := widget.NewLabel("WingnutTech LED Controller Updater")

	verSelect = widget.NewSelect([]string{}, func(value string) {
		updateLayouts(value)
	})
	verSelect.PlaceHolder = "(Select a version)"

	layoutSelect = widget.NewSelect([]string{}, func(value string) {
		if value == "-Custom-" {
			vboxCustom.Show()
			resizeMainWindow()
			ready.SelectionExists = true
		} else {
			vboxCustom.Hide()
			resizeMainWindow()
			if _, ok := releases[verSelect.Selected][value]; ok {
				ready.SelectionExists = true
			} else {
				ready.SelectionExists = false
			}
		}
		checkReady()
	})
	layoutSelect.PlaceHolder = "(Select a layout)"

	portSelect = widget.NewSelect([]string{}, func(value string) {
		port = value
		if port != "" {
			ready.Port = true
		} else {
			ready.Port = false
		}
		checkReady()
	})

	portRefreshBtn := widget.NewButton("Refresh COM ports", func() {
		getPorts()
	})

	flashBtn = widget.NewButton("Flash Firmware", func() {
		if layoutSelect.Selected == "-Custom-" {
			go compileAndFlash(verSelect.Selected)
		} else {
			go downloadAndFlash(verSelect.Selected, layoutSelect.Selected)
		}
	})
	flashBtn.Disable()

	// CH340 driver button on windows only
	driverBtn := widget.NewButton("CH340 Drivers", func() {
		go installCH340()
	})
	if runtime.GOOS != "windows" && false {
		driverBtn.Hide()
	}

	return container.NewVBox(
		topLabel,
		driverBtn,
		verSelect,
		layoutSelect,
		container.NewGridWithColumns(2,
			container.NewVBox(
				portSelect,
				portRefreshBtn,
			),
			flashBtn,
		),
	)
}

func makeCustomSection() *fyne.Container {
	wingLEDLabel := widget.NewLabel("Wing: ")
	wingRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		customConfig.WingRev = checked
	})
	wingLEDSlider := widget.NewSlider(0, 50)
	wingLEDSlider.OnChanged = func(value float64) {
		wingLEDLabel.SetText("Wing: " + fmt.Sprint(value))
		customConfig.WingLEDs = int(value)
	}
	wingLEDSlider.SetValue(19)

	noseLEDLabel := widget.NewLabel("Nose: ")
	noseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		customConfig.NoseRev = checked
	})
	noseLEDSlider := widget.NewSlider(0, 50)
	noseLEDSlider.OnChanged = func(value float64) {
		noseLEDLabel.SetText("Nose: " + fmt.Sprint(value))
		customConfig.NoseLEDs = int(value)
	}
	noseRevCheck.SetChecked(true)
	noseLEDSlider.SetValue(6)

	fuseLEDLabel := widget.NewLabel("Fuse: ")
	fuseRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		customConfig.FuseRev = checked
	})
	fuseLEDSlider := widget.NewSlider(0, 50)
	fuseLEDSlider.OnChanged = func(value float64) {
		fuseLEDLabel.SetText("Fuse: " + fmt.Sprint(value))
		customConfig.FuseLEDs = int(value)
	}
	fuseLEDSlider.SetValue(13)

	tailLEDLabel := widget.NewLabel("Tail: ")
	tailRevCheck := widget.NewCheck("Reversed?", func(checked bool) {
		customConfig.TailRev = checked
	})
	tailLEDSlider := widget.NewSlider(0, 50)
	tailLEDSlider.OnChanged = func(value float64) {
		tailLEDLabel.SetText("Tail: " + fmt.Sprint(value))
		customConfig.TailLEDs = int(value)
	}
	tailLEDSlider.SetValue(4)

	NavLEDLabel := widget.NewLabel("Nav LEDs: ")
	NavLEDSlider := widget.NewSlider(0, 50)
	NavLEDSlider.OnChanged = func(value float64) {
		NavLEDLabel.SetText("Nav LEDs: " + fmt.Sprint(value))
		customConfig.WingNavLEDs = int(value)
	}
	NavLEDSlider.SetValue(7)

	return container.NewVBox(
		container.NewGridWithColumns(2,
			wingLEDLabel, wingRevCheck,
		),
		wingLEDSlider,
		container.NewGridWithColumns(2,
			noseLEDLabel, noseRevCheck,
		),
		noseLEDSlider,
		container.NewGridWithColumns(2,
			fuseLEDLabel, fuseRevCheck,
		),
		fuseLEDSlider,
		container.NewGridWithColumns(2,
			tailLEDLabel, tailRevCheck,
		),
		tailLEDSlider,

		NavLEDLabel,
		NavLEDSlider,
	)
}

func resizeMainWindow() {
	mainWindow.Resize(fyne.Size{
		Width:  mainWinWidth,
		Height: mainWindow.Content().MinSize().Height + 10,
	})
}

func setStatus(s string) {
	statusLabel.SetText(s)
}

func errorPopup(s string) {
	popup := mainApp.NewWindow("Error")
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

func checkReady() {
	if ready.LibrariesInstalled &&
		ready.SelectionExists &&
		ready.Port &&
		ready.NotFlashing {
		flashBtn.Enable()
	} else {
		flashBtn.Disable()
	}
}

func getPorts() {
	port = ""
	ports := []string{}

	res, _ := board.List(inst.Id)
	for _, p := range res {
		ports = append(ports, p.Address)
		if strings.HasSuffix(p.ProtocolLabel, "(USB)") {
			port = p.Address
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i] > ports[j]
	})

	portSelect.Options = ports
	portSelect.SetSelected(port)
}

func checkLibraries() {
	r := true
	for _, libName := range neededLibraries {
		if err := lib.LibraryInstall(context.Background(), &rpc.LibraryInstallRequest{
			Instance: inst,
			Name:     libName,
		}, output.NewNullDownloadProgressCB(), output.NewNullTaskProgressCB()); err != nil {
			errorPopup(err.Error())
			r = false
		}
	}
	ready.LibrariesInstalled = r
	checkReady()
}

func updateReleases() {
	resp, err := http.Get(API_URL)
	if err != nil {
		errorPopup("Unable to get releases")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorPopup("Unable to get releases")
		return
	}

	all_releases := Releases{}
	err = json.Unmarshal(body, &all_releases)
	if err != nil {
		errorPopup("Unable to parse releases")
		return
	}

	for _, release := range all_releases {
		releases[release.Name] = map[string]string{}
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, ".hex") {
				releases[release.Name][asset.Name] = asset.BrowserDownloadURL
			}
		}
	}
	updateVersions()
}

func updateVersions() {
	o := []string{}
	for k := range releases {
		o = append(o, k)
	}
	sort.Slice(o, func(i, j int) bool {
		return o[i] > o[j]
	})
	verSelect.Options = o
	verSelect.SetSelectedIndex(0)
}

func updateLayouts(v string) {
	prevSelection := layoutSelect.Selected

	o := []string{}
	for k := range releases[v] {
		o = append(o, k)
	}
	sort.Slice(o, func(i, j int) bool {
		return o[i] > o[j]
	})
	o = append(o, "-Custom-")

	layoutSelect.Options = o
	if prevSelection == "-Custom-" {
		layoutSelect.SetSelected("-Custom-")
	} else {
		layoutSelect.SetSelectedIndex(0)
	}
}

func downloadAndFlash(v string, h string) {
	ready.NotFlashing = false
	checkReady()

	defer func() {
		ready.NotFlashing = true
		checkReady()
	}()

	hexFile := filepath.Join(tmpPath, h)
	if _, err := os.Stat(hexFile); err != nil {
		if err := downloadFile(hexFile, releases[v][h]); err != nil {
			errorPopup(err.Error())
			return
		}
	}
	if err := flashHex(hexFile); err != nil {
		errorPopup(err.Error())
	}
}

func compileAndFlash(v string) {
	ready.NotFlashing = false
	checkReady()

	defer func() {
		ready.NotFlashing = true
		checkReady()
	}()

	newFolder := filepath.Join(tmpPath, v)
	configFile := filepath.Join(newFolder, "config.h")
	exportDir := filepath.Join(newFolder, "build")

	if _, err := os.Stat(newFolder); err != nil {
		setStatus("Downloading " + v)
		zipName := v + ".zip"
		zipFile := filepath.Join(tmpPath, zipName)
		zipUrl := ZIP_URL_PREFIX + zipName

		err := downloadFile(zipFile, zipUrl)
		if err != nil {
			errorPopup(err.Error())
			return
		}

		fileNames, err := unzipFile(zipFile, tmpPath)
		if err != nil {
			errorPopup(err.Error())
			return
		}
		// assume first element is parent directory
		os.Rename(fileNames[0], newFolder)
		os.Rename(
			filepath.Join(newFolder, "FT-Night-Radian-LED-Controller.ino"),
			filepath.Join(newFolder, v+".ino"),
		)
	}

	setStatus("Compiling custom firmware")
	err := os.WriteFile(configFile, generateCustomConfig(), os.ModePerm)
	if err != nil {
		errorPopup(err.Error())
		return
	}
	if _, err := compile.Compile(context.Background(), &rpc.CompileRequest{
		Instance:   inst,
		Fqbn:       FQBN,
		SketchPath: newFolder,
		ExportDir:  exportDir,
	}, os.Stdout, os.Stderr, false); err != nil {
		fmt.Println(err)
	}

	setStatus("Flashing custom firmware")
	err = flashHex(filepath.Join(exportDir, v+".ino.hex"))
	if err != nil {
		errorPopup(err.Error())
	}

	ready.NotFlashing = true
	checkReady()
}

func flashHex(hexFile string) error {
	setStatus("Flashing " + filepath.Base(hexFile))
	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   inst,
		Fqbn:       FQBN,
		SketchPath: tmpPath,
		Port:       port,
		ImportFile: hexFile,
	}, io.Discard, io.Discard); err != nil { // os.Stdout, os.Stderr
		return err
	}

	setStatus("Done!")
	return nil
}

func downloadFile(filename string, url string) error {
	setStatus("Downloading " + filepath.Base(filename))
	defer setStatus("")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error downloading file")
	}

	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func unzipFile(filename string, dest string) ([]string, error) {
	setStatus("Unzipping " + filepath.Base(filename))
	defer setStatus("")

	fileNames := []string{}

	r, err := zip.OpenReader(filename)
	if err != nil {
		return fileNames, err
	}

	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fileNames, fmt.Errorf("%s: illegal filepath", fpath)
		}

		fileNames = append(fileNames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fileNames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fileNames, err
		}

		rc, err := f.Open()
		if err != nil {
			return fileNames, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fileNames, err
		}
	}

	return fileNames, nil
}

func installCH340() {
	exeFile := filepath.Join(tmpPath, "CH34x_Install_Windows_v3_4.EXE")
	zipFile := filepath.Join(tmpPath, "ch340.zip")

	if _, err := os.Stat(exeFile); err != nil {
		if _, err := os.Stat(zipFile); err != nil {
			err := downloadFile(zipFile, CH340_URL)
			if err != nil {
				errorPopup(err.Error())
				return
			}
		}

		_, err = unzipFile(zipFile, tmpPath)
		if err != nil {
			errorPopup(err.Error())
			return
		}
	}

	cmd := exec.Command(filepath.Join(tmpPath, "CH34x_Install_Windows_v3_4.EXE"))
	err := cmd.Start()
	if err != nil {
		errorPopup(err.Error())
	}
}

func generateCustomConfig() []byte {
	return []byte(fmt.Sprintf(
		"// number of LEDs in specific strings\n"+
			"#define WING_LEDS %d // total wing LEDs\n"+
			"#define NOSE_LEDS %d // total nose LEDs\n"+
			"#define FUSE_LEDS %d // total fuselage LEDs\n"+
			"#define TAIL_LEDS %d // total tail LEDs\n"+
			"\n"+
			"// strings reversed?\n"+
			"#define WING_REV %t\n"+
			"#define NOSE_REV %t\n"+
			"#define FUSE_REV %t\n"+
			"#define TAIL_REV %t\n"+
			"\n"+
			"#define NOSE_FUSE_JOINED true // are the nose and fuse strings joined?\n"+
			"#define WING_NAV_LEDS %d // wing LEDs that are navlights\n"+
			"\n"+
			"#define TMP_BRIGHTNESS 175\n",
		customConfig.WingLEDs, customConfig.NoseLEDs,
		customConfig.FuseLEDs, customConfig.TailLEDs,
		customConfig.WingRev, customConfig.NoseRev,
		customConfig.FuseRev, customConfig.TailRev,
		customConfig.WingNavLEDs,
	))
}
