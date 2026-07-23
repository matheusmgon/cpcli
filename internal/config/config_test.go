package config

import (
	"path/filepath"
	"testing"
)

func TestDirPrefersXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg-base")
	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	want := filepath.Join("/xdg-base", "cpcli")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDirFallsBackToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/synthetic-user")
	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	want := filepath.Join("/home/synthetic-user", ".config", "cpcli")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolvePath(t *testing.T) {
	originalWD = "/original/cwd"

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty path is left as-is", "", ""},
		{"absolute path is left as-is", "/already/absolute/file.json", "/already/absolute/file.json"},
		{"relative path is joined against the original cwd", "payload.json", "/original/cwd/payload.json"},
		{"relative path with subdirectory", "sub/payload.json", "/original/cwd/sub/payload.json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolvePath(tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
