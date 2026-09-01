// Package config resolves non-secret Buzz integration settings for a camp.
package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileName is the camp-relative bind file path (no secrets).
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
	if err := ensureSafeConfigPath(campaignRoot); err != nil {
		return Config{}, path, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, path, nil
		}
		return Config{}, path, fmt.Errorf("read %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parse %q: %w", path, err)
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
// The Buzz CLI documents BUZZ_RELAY_URL as an HTTP base (default
// http://localhost:3000) and normalizes WebSocket forms to HTTP forms.
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
		return "", fmt.Errorf("camp root required")
	}
	cfg.RelayURL = NormalizeRelayURL(cfg.RelayURL)
	path := filepath.Join(campaignRoot, FileName)
	if err := ensureSafeConfigPath(campaignRoot); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create integrations dir: %w", err)
	}
	if err := ensureSafeConfigPath(campaignRoot); err != nil {
		return "", err
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}
	header := []byte("# Non-secret Buzz bind for camp-buzz. Never put private keys here.\n")
	if err := writeFileAtomically(path, append(header, data...), os.Rename); err != nil {
		return "", fmt.Errorf("write %q: %w", path, err)
	}
	return path, nil
}

func ensureSafeConfigPath(campaignRoot string) error {
	current := campaignRoot
	for _, part := range strings.Split(FileName, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect config path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlinked config path component %q", part)
		}
	}
	return nil
}

func writeFileAtomically(path string, data []byte, rename func(string, string) error) error {
	mode, err := configFileMode(path)
	if err != nil {
		return err
	}
	temporary, err := writeTemporary(filepath.Dir(path), data, mode)
	if err != nil {
		return err
	}
	defer os.Remove(temporary)
	if _, err := configFileMode(path); err != nil {
		return err
	}
	if err := rename(temporary, path); err != nil {
		return fmt.Errorf("replace config atomically: %w", err)
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return fmt.Errorf("config replaced, but directory durability could not be confirmed: %w", err)
	}
	return nil
}

func configFileMode(path string) (fs.FileMode, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return 0o644, nil
	}
	if err != nil {
		return 0, fmt.Errorf("inspect existing config: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, fmt.Errorf("refusing symlinked config target")
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("config target is not a regular file")
	}
	return info.Mode().Perm(), nil
}

func writeTemporary(directory string, data []byte, mode fs.FileMode) (string, error) {
	file, err := os.CreateTemp(directory, ".buzz.yaml.tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary config: %w", err)
	}
	path := file.Name()
	complete := false
	defer func() {
		if !complete {
			file.Close()
			os.Remove(path)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return "", fmt.Errorf("set temporary config permissions: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("write temporary config: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync temporary config: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close temporary config: %w", err)
	}
	complete = true
	return path, nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync config directory: %w", err)
	}
	return nil
}
