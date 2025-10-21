//go:build android

package main

import (
	"os"
)

// android stops network of background app, so exit
func lifecycle(g GUI) {
	lifecycle := g.app.Lifecycle()
	lifecycle.SetOnExitedForeground(func() {
		os.Exit(0)
	})
}
