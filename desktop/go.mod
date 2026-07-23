// The desktop (Wails) app is a SEPARATE Go module on purpose: it pulls in the
// Wails runtime, which needs CGO + libwebkit2gtk at build time on Linux.
// Keeping it out of the root cpcli module means `go build ./...` at the repo
// root stays buildable everywhere, and only this module needs the GUI deps.
//
// It reuses the shared UI facade from the core module via the replace below.
// After installing the deps (see README.md), run `go mod tidy` here to pin the
// exact Wails version, then `wails dev` / `wails build`.
module cpcli/desktop

go 1.25.11

require (
	cpcli v0.0.0
	github.com/wailsapp/wails/v2 v2.10.1
)

replace cpcli => ../
