//go:build integration

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// End-to-end against a fake buzz binary (no real relay).
func TestIntegrationFixtureFlow(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repo := wd
	for d := wd; d != "/" && d != "."; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			// prefer camp-buzz module root
			data, _ := os.ReadFile(filepath.Join(d, "go.mod"))
			if strings.Contains(string(data), "camp-buzz") {
				repo = d
				break
			}
		}
	}

	script := filepath.Join(repo, "scripts/vhs-fixture.sh")
	if _, err := os.Stat(script); err != nil {
		t.Skip("fixture script not found:", script)
	}

	root := t.TempDir()
	cmd := exec.Command("bash", script)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "CAMP_BUZZ_VHS_ROOT="+root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vhs-fixture: %v\n%s", err, out)
	}
	fixRoot := strings.TrimSpace(string(out))
	bin := filepath.Join(fixRoot, "bin", "camp-buzz")

	run := func(args ...string) (string, error) {
		c := exec.Command(bin, args...)
		c.Env = append(os.Environ(),
			"HOME="+filepath.Join(fixRoot, "home"),
			"PATH="+filepath.Join(fixRoot, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
			"CAMP_ROOT="+filepath.Join(fixRoot, "campaign"),
			"CAMP_BUZZ_ROOT="+filepath.Join(fixRoot, "home", ".obey", "plugins", "camp-buzz"),
			"BUZZ_PRIVATE_KEY=demo-nsec-not-real",
			"BUZZ_FAKE_LOG="+filepath.Join(fixRoot, "fake-buzz.log"),
		)
		var buf bytes.Buffer
		c.Stdout = &buf
		c.Stderr = &buf
		err := c.Run()
		return buf.String(), err
	}

	if out, err := run("version"); err != nil {
		t.Fatalf("version: %v\n%s", err, out)
	}

	if out, err := run("bind",
		"--channel", "44444444-4444-4444-8444-444444444444",
		"--relay", "ws://localhost:3000",
		"--festival", "ITEST1",
	); err != nil {
		t.Fatalf("bind: %v\n%s", err, out)
	}
	if out, err := run("bind", "--festival", "ITEST2"); err != nil {
		t.Fatalf("merge existing bind: %v\n%s", err, out)
	}

	outStr, err := run("doctor")
	if err != nil {
		t.Fatalf("doctor expected ready: %v\n%s", err, outStr)
	}
	if !strings.Contains(outStr, "status: ready") {
		t.Fatalf("doctor output:\n%s", outStr)
	}

	outStr, err = run("post", "-m", "integration test body", "--task", "FEST-itest", "--gate", "pass")
	if err != nil {
		t.Fatalf("post: %v\n%s", err, outStr)
	}
	if !strings.Contains(outStr, "posted") {
		t.Fatalf("post output:\n%s", outStr)
	}

	logPath := filepath.Join(fixRoot, "fake-buzz.log")
	logBody, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(logBody)
	for _, want := range []string{"integration test body", "festival: ITEST2", "task: FEST-itest", "gate: pass"} {
		if !strings.Contains(s, want) {
			t.Fatalf("fake buzz log missing %q:\n%s", want, s)
		}
	}
}
