package session

import (
	"errors"
	"testing"
)

func TestSaveLoadClearRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	const profile = "test-profile"

	if _, err := Load(profile); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("expected ErrNotLoggedIn before any save, got %v", err)
	}

	want := &Session{
		Server:     "lab.example",
		Port:       443,
		Sid:        "synthetic-session-id",
		ApiVersion: "1.9",
		User:       "admin",
	}
	if err := Save(profile, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(profile)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != *want {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	if err := Clear(profile); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := Load(profile); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("expected ErrNotLoggedIn after Clear, got %v", err)
	}

	// Clearing an already-absent session must not error.
	if err := Clear(profile); err != nil {
		t.Fatalf("Clear on already-absent session: %v", err)
	}
}

func TestRejectsPathTraversalInProfileName(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	malicious := []string{
		"../../etc/passwd",
		"../sibling",
		"a/b",
		"",
	}
	for _, profile := range malicious {
		if _, err := filePath(profile); err == nil {
			t.Fatalf("filePath(%q): expected an error, got none", profile)
		}
	}
}
