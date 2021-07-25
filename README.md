# LED Controller Updater

This is a companion program to the [WingnutTech LED Controller](https://github.com/wingnut-tech/FT-Night-Radian-LED-Controller) firmware for R/C planes.

I wanted to make an easy way for people to update their controllers, and even compile custom firmware, without having to mess around with installing and configuring the full Arduino IDE.

This project is only possible due to the [Fyne](https://fyne.io/) GUI library for Go, and the [arduino-cli](https://github.com/arduino/arduino-cli) project.  
Because of them, I was able to create a truly cross-platform utility that works the same on Windows, Mac, and Linux.

## Getting the app

If you'd like to download and compile this program, go* for it. Just clone the repo and run a `go build`.  
Keep in mind you will need your OS's C/C++ compiler and a few development libraries due to some CGO dependencies.

But the easiest way to get going is to just download the latest pre-compiled executable for your OS from the [Releases](https://github.com/reyemxela/LEDControllerUpdater/releases) page.  
There's no installation, just run the file. The app takes care of downloading a few arduino core files and libraries in the background on first launch.



<sub>\**I'm not sorry</sub>*
