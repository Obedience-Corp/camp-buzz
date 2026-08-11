package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeRelayURL(t *testing.T) {
	cases := map[string]string{
		"ws://localhost:3000":   "http://localhost:3000",
		"wss://relay.example":   "https://relay.example",
		"http://localhost:3000": "http://localhost:3000",
		"https://relay.example": "https://relay.example",
		"":                      "",
	}
	for in, want := range cases {
		if got := NormalizeRelayURL(in); got != want {
			t.Fatalf("NormalizeRelayURL(%q)=%q want %q", in, got, want)
		}
	}
}

func TestWriteAndResolve(t *testing.T) {
	t.Setenv("BUZZ_RELAY_URL", "")
	t.Setenv("BUZZ_CHANNEL", "")
	root := t.TempDir()
	path, err := Write(root, Config{
		RelayURL:   "ws://localhost:3000",
		ChannelID:  "chan-1",
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
	// file stores ws://; Resolve normalizes to http:// for the buzz CLI
	if cfg.ChannelID != "chan-1" || cfg.RelayURL != "http://localhost:3000" {
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
	// ws:// from env should normalize to http://
	t.Setenv("BUZZ_RELAY_URL", "ws://relay.example:3000")
	cfg, _, err = Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RelayURL != "http://relay.example:3000" {
		t.Fatalf("expected ws→http normalize, got %q", cfg.RelayURL)
	}
	// path shape
	want := filepath.Join(root, FileName)
	if path != want {
		t.Fatalf("path = %q want %q", path, want)
	}
}
