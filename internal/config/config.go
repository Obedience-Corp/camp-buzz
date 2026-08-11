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

// Resolve loads bind file (if present) then overlays env (env wins).
func Resolve(campaignRoot string) (Config, string, error) {
	var cfg Config
	path := ""
	if campaignRoot != "" {
		path = filepath.Join(campaignRoot, FileName)
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return Config{}, path, fmt.Errorf("parse %s: %w", path, err)
			}
		} else if !os.IsNotExist(err) {
			return Config{}, path, fmt.Errorf("read %s: %w", path, err)
		}
	}
	if v := os.Getenv("BUZZ_RELAY_URL"); v != "" {
		cfg.RelayURL = v
	}
	if v := os.Getenv("BUZZ_CHANNEL"); v != "" {
		cfg.ChannelID = v
	}
	return cfg, path, nil
}

// Write upserts the bind file under campaignRoot.
func Write(campaignRoot string, cfg Config) (string, error) {
	if campaignRoot == "" {
		return "", fmt.Errorf("campaign root required")
	}
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
