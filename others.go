//go:build !android

package main

import "fyne.io/fyne/v2"

func lifecycle(g GUI) {
	g.window.Resize(fyne.NewSize(400, 700))
}
