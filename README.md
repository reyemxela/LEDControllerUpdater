# LED Controller Updater

# This project is now deprecated
With the release of our v2 board and the complete overhaul of the code (which supports both v1 and v2 boards), this updater app is no longer needed in its current state.

However, there wasn't a good way to update either the firmware or this app without everything breaking. So both this repo and the original firmware repo will stay as-is and be archived, with the new code going live in a [new repo](https://github.com/wingnut-tech/LEDControllerV2). Follow the instructions there to update your board to the new v2 firmware.

---

This is a companion program to the [WingnutTech LED Controller](https://github.com/wingnut-tech/LEDController) firmware for R/C planes.

I wanted to make an easy way for people to update their controllers, and even compile custom firmware, without having to mess around with installing and configuring the full Arduino IDE.

This project is only possible due to the [Fyne](https://fyne.io/) GUI library for Go, and the [arduino-cli](https://github.com/arduino/arduino-cli) project.  
Because of them, I was able to create a truly cross-platform utility that works the same on Windows, Mac, and Linux.

## Getting the app

If you'd like to download and compile this program, go* for it. Just clone the repo and run a `go build`.  
Keep in mind you will need your OS's C/C++ compiler and a few development libraries due to some CGO dependencies.

But the easiest way to get going is to just download the latest pre-compiled executable for your OS from the [Releases](https://github.com/reyemxela/LEDControllerUpdater/releases) page.  
There's no installation, just run the file. The app takes care of downloading a few arduino core files and libraries in the background on first launch.



<sub>\**I'm not sorry</sub>*
