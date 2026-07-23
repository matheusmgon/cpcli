// Package session persists the Check Point Management API session (sid)
// returned by "login" so that each cpcli invocation (a separate OS process)
// can reuse it instead of logging in again. It never stores credentials —
// only the session id and the connection details needed to keep using it.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cpcli/internal/config"
)

var ErrNotLoggedIn = errors.New("não autenticado — rode `cpcli login` primeiro (use --profile se estiver gerenciando mais de um servidor)")

// Session is the persisted state of a logged-in profile.
type Session struct {
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Sid        string `json:"sid"`
	ApiVersion string `json:"api_version,omitempty"`
	Domain     string `json:"domain,omitempty"`
	User       string `json:"user,omitempty"`
	ReadOnly   bool   `json:"read_only"`
	// Insecure records whether login was done with --insecure (TLS
	// fingerprint verification skipped). It must be carried into every
	// later command's client for the same reason it was set at login: an
	// insecure login never pins a fingerprint, so if later commands built
	// their client with verification back on, they'd hit the SDK's
	// interactive fingerprint prompt (or fail outright in non-interactive
	// use) on every single call.
	Insecure bool `json:"insecure"`
}

// filePath returns the on-disk path for a profile's session file. profile
// comes straight from the user-facing --profile flag, so it's constrained
// to a single path segment to keep Load/Save/Clear from being pointed at an
// arbitrary file via a value like "../../etc/passwd".
func filePath(profile string) (string, error) {
	if profile == "" || profile != filepath.Base(profile) {
		return "", fmt.Errorf("nome de perfil inválido: %q", profile)
	}
	dir, err := config.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve config directory: %w", err)
	}
	return filepath.Join(dir, fmt.Sprintf("session-%s.json", profile)), nil
}

// Load reads the persisted session for the given profile. It returns
// ErrNotLoggedIn if no session was ever saved for that profile.
func Load(profile string) (*Session, error) {
	p, err := filePath(profile)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, fmt.Errorf("failed to read session file %s: %w", p, err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("session file %s is corrupted: %w", p, err)
	}
	return &s, nil
}

// Save persists the session for the given profile, replacing any previous one.
func Save(profile string, s *Session) error {
	p, err := filePath(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("failed to write session file %s: %w", p, err)
	}
	return nil
}

// Clear removes the persisted session for the given profile, if any.
func Clear(profile string) error {
	p, err := filePath(profile)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file %s: %w", p, err)
	}
	return nil
}
