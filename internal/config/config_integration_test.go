//go:build integration

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndResolve(t *testing.T) {
	t.Setenv("BUZZ_RELAY_URL", "")
	t.Setenv("BUZZ_CHANNEL", "")
	root := t.TempDir()
	path, err := Write(root, Config{
		RelayURL:   "ws://localhost:3000",
		ChannelID:  "33333333-3333-4333-8333-333333333333",
		FestivalID: "CI0009",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, path, 0o644)
	assertResolvedConfig(t, root)
	assertEnvironmentOverlay(t, root)
	if want := filepath.Join(root, FileName); path != want {
		t.Fatalf("path = %q want %q", path, want)
	}
}

func TestWritePreservesExistingPermissions(t *testing.T) {
	root := t.TempDir()
	path, err := Write(root, Config{ChannelID: "33333333-3333-4333-8333-333333333333"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, Config{ChannelID: "44444444-4444-4444-8444-444444444444"}); err != nil {
		t.Fatal(err)
	}
	assertMode(t, path, 0o600)
}

func TestInterruptedWritePreservesExistingConfig(t *testing.T) {
	root := t.TempDir()
	path, err := Write(root, Config{FestivalID: "original"})
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	renameFailure := errors.New("injected rename failure")
	err = writeFileAtomically(path, []byte("replacement"), func(string, string) error { return renameFailure })
	if !errors.Is(err, renameFailure) {
		t.Fatalf("error = %v", err)
	}
	assertFileContent(t, path, string(original))
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".buzz.yaml.tmp-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files after failure = %v, error = %v", matches, err)
	}
}

func TestWriteRefusesSymlinkedTarget(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("outside-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, FileName)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, Config{FestivalID: "replacement"}); err == nil || !strings.Contains(err.Error(), "symlinked") {
		t.Fatalf("error = %v", err)
	}
	assertFileContent(t, outside, "outside-original")
}

func TestWriteRefusesSymlinkedParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	campaignDir := filepath.Join(root, ".campaign")
	if err := os.Mkdir(campaignDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(campaignDir, "integrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, Config{FestivalID: "replacement"}); err == nil || !strings.Contains(err.Error(), "symlinked") {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "buzz.yaml")); !os.IsNotExist(err) {
		t.Fatalf("outside target was created: %v", err)
	}
}

func TestWriteReportsDirectoryTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, FileName)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, Config{}); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %v", err)
	}
}

func assertResolvedConfig(t *testing.T, root string) {
	t.Helper()
	cfg, _, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ChannelID != "33333333-3333-4333-8333-333333333333" || cfg.RelayURL != "http://localhost:3000" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func assertEnvironmentOverlay(t *testing.T, root string) {
	t.Helper()
	t.Setenv("BUZZ_CHANNEL", "44444444-4444-4444-8444-444444444444")
	t.Setenv("BUZZ_RELAY_URL", "ws://relay.example:3000")
	cfg, _, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ChannelID != "44444444-4444-4444-8444-444444444444" || cfg.RelayURL != "http://relay.example:3000" {
		t.Fatalf("environment overlay failed: %+v", cfg)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("mode = %o, want %o", got, want)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}
