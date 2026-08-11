// Package config resolves non-secret Buzz integration settings for a campaign.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the campaign-relative bind file (no secrets).
const FileName = ".campaign/integrations/buzz.yaml"

// Config is the non-secret bind configuration.
type Config struct {
	RelayURL     string `yaml:"relay_url,omitempty"`
	ChannelID    string `yaml:"channel_id,omitempty"`
	FestivalID   string `yaml:"festival_id,omitempty"`
	FestivalPath string `yaml:"festival_path,omitempty"`
}

// LoadFile reads the bind file only (no env overlay). Missing file returns empty Config.
func LoadFile(campaignRoot string) (Config, string, error) {
	var cfg Config
	if campaignRoot == "" {
		return cfg, "", nil
	}
	path := filepath.Join(campaignRoot, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, path, nil
		}
		return Config{}, path, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, path, nil
}

// Resolve loads bind file (if present) then overlays env (env wins).
// Relay URLs are normalized for the buzz CLI (HTTP base URL, not WebSocket).
func Resolve(campaignRoot string) (Config, string, error) {
	cfg, path, err := LoadFile(campaignRoot)
	if err != nil {
		return Config{}, path, err
	}
	if v := os.Getenv("BUZZ_RELAY_URL"); v != "" {
		cfg.RelayURL = v
	}
	if v := os.Getenv("BUZZ_CHANNEL"); v != "" {
		cfg.ChannelID = v
	}
	cfg.RelayURL = NormalizeRelayURL(cfg.RelayURL)
	return cfg, path, nil
}

// NormalizeRelayURL converts ws:// / wss:// to http:// / https:// for buzz CLI.
// The real buzz CLI documents BUZZ_RELAY_URL as an HTTP base (default
// http://localhost:3000). WebSocket forms are accepted by some builds after
// normalization, but HTTP is the supported contract.
func NormalizeRelayURL(u string) string {
	switch {
	case len(u) >= 6 && u[:6] == "wss://":
		return "https://" + u[6:]
	case len(u) >= 5 && u[:5] == "ws://":
		return "http://" + u[5:]
	default:
		return u
	}
}

// Write upserts the bind file under campaignRoot.
func Write(campaignRoot string, cfg Config) (string, error) {
	if campaignRoot == "" {
		return "", fmt.Errorf("campaign root required")
	}
	cfg.RelayURL = NormalizeRelayURL(cfg.RelayURL)
	path := filepath.Join(campaignRoot, FileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create integrations dir: %w", err)
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}
	header := []byte("# Non-secret Buzz bind for camp-buzz. Never put private keys here.\n")
	if err := os.WriteFile(path, append(header, data...), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}
