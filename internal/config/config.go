// Package config resolves where cpcli keeps its local state (sessions and
// the TLS fingerprint trust store).
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var originalWD string

// Init prepares the config directory and must be called once at startup,
// before any Check Point API client is created.
//
// The upstream Check Point Go SDK persists trusted TLS fingerprints to a
// file named "fingerprints.json" resolved relative to the process's current
// directory (it has no option to configure an absolute path). Without this
// chdir, the trust store would fragment per invocation directory and users
// would be re-prompted to accept the server's fingerprint every time they
// ran cpcli from a different folder. Chdir'ing into a stable config
// directory at startup keeps a single trust store regardless of where cpcli
// is invoked from.
func Init() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine current directory: %w", err)
	}
	originalWD = wd

	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to switch into config directory %s: %w", dir, err)
	}
	return nil
}

// Dir returns the directory cpcli stores its state in.
func Dir() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "cpcli"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "cpcli"), nil
}

// ResolvePath resolves a user-supplied relative path against the directory
// cpcli was originally invoked from, not the config directory Init chdir'd
// into. Any command flag that accepts a filesystem path from the user must
// pass it through this before opening it.
func ResolvePath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(originalWD, p)
}
