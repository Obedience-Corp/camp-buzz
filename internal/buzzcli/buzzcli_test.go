package buzzcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSendMessageWithFakeBuzz(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "buzz")
	// minimal fake that accepts messages send
	script := `#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == "messages" && "$2" == "send" ]]; then
  cat >/dev/null
  echo '{"ok":true}'
  exit 0
fi
exit 1
`
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")

	p, err := LookPath()
	if err != nil {
		t.Fatal(err)
	}
	if p != fake && filepath.Base(p) != "buzz" {
		t.Fatalf("LookPath = %q", p)
	}
	if !HasPrivateKey() {
		t.Fatal("expected private key set")
	}
	if err := SendMessage("chan-1", "hello\n\n---\nfestival: X\ntask: -\npath: -\ngate: n/a\n---\n", "ws://localhost:3000"); err != nil {
		t.Fatal(err)
	}
}

func TestSendMessageRequiresKey(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "")
	// ensure empty
	_ = os.Unsetenv("BUZZ_PRIVATE_KEY")
	err := SendMessage("c", "body", "")
	if err == nil {
		t.Fatal("expected error without key")
	}
}
