## [v1.2.0] - 2022-10-26
Complete project overhaul

The back-end logic and functionality has been separated from the frontend UI, which allows different frontends to be developed and used.

So now in addition to the normal GUI version, we have an interactive CLI version of the app, which should work for those with older computers that might have OpenGL issues.

I was also able to make most areas of the code a lot less inter-dependant on others, which should help with maintainability and make it easier to add new features or even other frontends.

### Added
- GUI/CLI versions of the app

### Changed
- Everything
- All of it
- But seriously, in redesigning the GUI from the ground up, I did a bit of tweaking of the layout. The Custom section now opens to the right instead of below, which helps the main window not get so tall.


## [v1.1.2] - 2022-09-25
### Changed
- Adjustable logging levels, mainly for my benefit when developing. Could be useful in the future for diagnosing other issues that people might have.

### Fixed
- Improved bootloader detection, to hopefully fix some cases where it was failing
- Fixed some auto-updater errors when the "tmp" folder is on a different drive

## [v1.1.1] - 2022-09-12
### Changed
- Nav LEDS are now dynamically limited to the Wing LED count
- Added Nose-Fuse join checkbox

### Fixed
- Window should center on resize correctly now
- Fixed a potential crash if checking for new firmware/app releases comes back empty

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