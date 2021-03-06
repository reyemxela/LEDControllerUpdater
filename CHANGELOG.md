## [v1.1.0] - 2021-09-26
### New firmware repository and structure
The firmware repo has been renamed from FT-Night-Radian-LED-Controller to just LEDController, and overhauled in the process. This new project structure required breaking changes to the Updater app.

The old firmware repo has been left in place so as not to break old versions of the updater, but will be archived.

### Added
- The app now has auto-update functionality on Windows and Linux. It will check for newer versions on start, and ask if you want to update. It will then automatically download the latest version, install it in-place, and relaunch. _Unfortunately due to Apple's sandboxing, this feature won't work on macOS. You will still get a pop-up alerting you to the new version, with a link to download it._

### Changed
- Point to new firmware repo and updated filenames.
- Added `LED_POWER` define to custom layout generation. There's a difference in power draw between the Night Radian and the strips we're using for the kits, so this allows us to specify the power draw per LED to keep power usage in check.
- The app now does a quick check to determine the bootloader version before flashing. In the past, the wait time could be quite long, depending on which bootloader the chip had.

## [v1.0.4] - 2021-09-20
### Changed
- Now specifying library versions. BMP280 v2.4.0 adds ~1KB to compiled hex size, and we're already tight.

## [v1.0.3] - 2021-09-01
### Added
- **Batch Mode toggle:** This is mainly for internal use, pressing ctrl+shift+b puts the app in "Batch Mode." This will auto-flash any new arduino that's plugged in, making it easier to flash a bunch of new boards when we're building them.

## [v1.0.2] - 2021-07-31
### Fixed
- The upload function now works correctly with both the atmega328 and atmega328old bootloaders. It now catches the specific error the old bootloader throws, and automatically switches to the atmega328old setting.

## [v1.0.1] - 2021-07-24
### Added
- Version number is now in the top label

### Changed
- `GetPorts()` now uses the arduino-cli `board.Watch()` function to automatically respond to COM port plug/unplug events. This also means the refresh button could go away.

### Fixed
- The initial updates and library/core checking are now done sequentially. I noticed a random error when running once, and it looked like there were some conflicts between the different functions trying to run at exactly the same time.

## [v1.0.0] - 2021-07-24
First "real" release, also where the support files will live.