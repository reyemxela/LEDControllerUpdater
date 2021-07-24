// +build !windows

package main

// these functions aren't needed on non-windows OSes
func (a *App) InstallCH340() {
}

func HideConsoleWindow() {
}
