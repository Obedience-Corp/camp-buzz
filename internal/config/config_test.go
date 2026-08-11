package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndResolve(t *testing.T) {
	root := t.TempDir()
	path, err := Write(root, Config{
		RelayURL:  "ws://localhost:3000",
		ChannelID: "chan-1",
		FestivalID: "CI0009",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ChannelID != "chan-1" || cfg.RelayURL != "ws://localhost:3000" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	// env wins
	t.Setenv("BUZZ_CHANNEL", "from-env")
	cfg, _, err = Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ChannelID != "from-env" {
		t.Fatalf("env should win, got %q", cfg.ChannelID)
	}
	// path shape
	want := filepath.Join(root, FileName)
	if path != want {
		t.Fatalf("path = %q want %q", path, want)
	}
}
