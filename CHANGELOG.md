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