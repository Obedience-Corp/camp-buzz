// Package assets resolves camp-buzz runtime asset root.
package assets

import (
	"os"
	"path/filepath"
)

// EnvRoot overrides the asset root.
const EnvRoot = "CAMP_BUZZ_ROOT"

// DefaultPluginDir is the Obedience plugin runtime location.
func DefaultPluginDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".obey", "plugins", "camp-buzz")
}

// Root returns the selected asset root and a short reason.
func Root() (string, string) {
	if v := os.Getenv(EnvRoot); v != "" {
		return v, "CAMP_BUZZ_ROOT"
	}
	if d := DefaultPluginDir(); d != "" {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d, "default ~/.obey/plugins/camp-buzz"
		}
	}
	// development: executable-adjacent or cwd assets/
	if exe, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(exe), "..", "assets")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand, "executable-adjacent assets"
		}
	}
	if st, err := os.Stat("assets"); err == nil && st.IsDir() {
		return "assets", "cwd assets/"
	}
	return DefaultPluginDir(), "default (may be missing)"
}
