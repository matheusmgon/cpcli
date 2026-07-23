// Command cpcli-desktop is a Wails-based desktop client for the Check Point
// Management API. It binds the shared UI facade (cpcli/service) so the same
// operations the CLI performs are available behind a graphical interface —
// the goal being a SmartConsole-like experience on Linux and macOS.
//
// Build/run requires the Wails toolchain and, on Linux, libwebkit2gtk; see
// README.md. This file is scaffolding validated by `wails build`, not by the
// core module's test suite.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"cpcli/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := service.New()

	err := wails.Run(&options.App{
		Title:            "cpcli — Check Point Management",
		Width:            1100,
		Height:           720,
		MinWidth:         820,
		MinHeight:        560,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		// Binding the facade exposes its methods to the frontend as
		// window.go.service.Service.<Method> (Promise-returning).
		Bind: []interface{}{svc},
	})
	if err != nil {
		println("erro ao iniciar o app:", err.Error())
	}
}
