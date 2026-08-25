//go:build integration

package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootPrefersEnvironmentOverride(t *testing.T) {
	t.Setenv(EnvRoot, "/managed/camp-buzz")
	root, reason := Root()
	if root != "/managed/camp-buzz" || reason != EnvRoot {
		t.Fatalf("Root() = %q, %q", root, reason)
	}
}

func TestRootUsesManagedPluginDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(EnvRoot, "")
	want := filepath.Join(home, ".obey", "plugins", "camp-buzz")
	if err := os.MkdirAll(want, 0o755); err != nil {
		t.Fatal(err)
	}
	root, reason := Root()
	if root != want || reason != "default ~/.obey/plugins/camp-buzz" {
		t.Fatalf("Root() = %q, %q", root, reason)
	}
}

func TestRootUsesDevelopmentAssetsFromWorkingDirectory(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(EnvRoot, "")
	if err := os.Mkdir(filepath.Join(work, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)
	root, reason := Root()
	if root != "assets" || reason != "cwd assets/" {
		t.Fatalf("Root() = %q, %q", root, reason)
	}
}
