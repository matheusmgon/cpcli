// Command cpcli is a command-line client for the Check Point Management API
// (the same API SmartConsole uses), for platforms without a native
// SmartConsole client (Linux, macOS).
package main

import (
	"fmt"
	"os"

	"cpcli/internal/cli"
	"cpcli/internal/config"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
