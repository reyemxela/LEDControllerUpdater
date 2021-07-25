#!/bin/bash

VERSION=$(cat main.go|grep "APP_VERSION" |cut -d '"' -f 2 |tr -d "v")
MAINDIR=$PWD

fyne-cross linux -app-version $VERSION
fyne-cross windows -console -app-version $VERSION
fyne-cross darwin -app-id LEDControllerUpdater -app-version $VERSION

cd $MAINDIR/fyne-cross/bin/linux-amd64/
rm $MAINDIR/releases/LEDControllerUpdater_linux.tar.gz
tar -czf $MAINDIR/releases/LEDControllerUpdater_linux.tar.gz LEDControllerUpdater

cd $MAINDIR/fyne-cross/dist/windows-amd64/
rm $MAINDIR/releases/LEDControllerUpdater_windows.zip
cp LEDControllerUpdater.exe.zip $MAINDIR/releases/LEDControllerUpdater_windows.zip

cd $MAINDIR/fyne-cross/dist/darwin-amd64/
rm $MAINDIR/releases/LEDControllerUpdater_mac.zip
zip -r $MAINDIR/releases/LEDControllerUpdater_mac.zip LEDControllerUpdater.app

cd $MAINDIR
