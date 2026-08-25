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

type cliFixture struct {
	root, bin string
	env       []string
}

const integrationChannel = "44444444-4444-4444-8444-444444444444"

func TestIntegrationFixtureFlow(t *testing.T) {
	fixture := newCLIFixture(t)
	mustRun(t, fixture, nil, "version")
	mustRun(t, fixture, nil, "bind", "--channel", integrationChannel, "--relay", "ws://localhost:3000", "--festival", "ITEST1")
	mustRun(t, fixture, nil, "bind", "--festival", "ITEST2")
	assertOutput(t, mustRun(t, fixture, nil, "doctor"), "status: ready")
	show := mustRun(t, fixture, nil, "show")
	assertOutput(t, show, "BUZZ_PRIVATE_KEY: [set]", "festival_id: ITEST2")
	if strings.Contains(show, "demo-nsec-not-real") {
		t.Fatal("show disclosed the private key")
	}
	hook := mustRun(t, fixture, nil, "hook-install")
	assertOutput(t, hook, "buzz_status:", "fail: open", "timeout: 30s")
	mustRun(t, fixture, nil, "post", "-m", "integration test body", "--task", "FEST-itest", "--gate", "pass")
	mustRun(t, fixture, nil, "post", "--from-hook")
	logBody, err := os.ReadFile(filepath.Join(fixture.root, "fake-buzz.log"))
	if err != nil {
		t.Fatal(err)
	}
	assertOutput(t, string(logBody), "integration test body", "Festival status update", "festival: ITEST2", "task: FEST-itest", "gate: pass")
}

func TestCommandFailuresAreActionable(t *testing.T) {
	fixture := newCLIFixture(t)
	unbound := filepath.Join(fixture.root, "unbound")
	if err := os.MkdirAll(filepath.Join(unbound, ".campaign"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := fixture.run([]string{"CAMP_ROOT=" + unbound}, "bind", "--festival", "FE1")
	assertFailure(t, out, err, "--channel is required")
	mustRun(t, fixture, nil, "bind", "--channel", integrationChannel, "--relay", "http://localhost:3000")
	out, err = fixture.run([]string{"BUZZ_PRIVATE_KEY="}, "doctor")
	assertFailure(t, out, err, "BUZZ_PRIVATE_KEY: NOT set", "status: not ready")
	out, err = fixture.run([]string{"BUZZ_PRIVATE_KEY="}, "post", "-m", "body")
	assertFailure(t, out, err, "BUZZ_PRIVATE_KEY is not set")
	badConfig := filepath.Join(fixture.root, "campaign", ".campaign", "integrations", "buzz.yaml")
	if err := os.WriteFile(badConfig, []byte("channel_id: [invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = fixture.run(nil, "show")
	assertFailure(t, out, err, "parse")
}

func newCLIFixture(t *testing.T) *cliFixture {
	t.Helper()
	repo, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	cmd := exec.Command("bash", filepath.Join(repo, "scripts", "vhs-fixture.sh"))
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "CAMP_BUZZ_VHS_ROOT="+root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vhs-fixture: %v\n%s", err, out)
	}
	fixtureRoot := strings.TrimSpace(string(out))
	return &cliFixture{
		root: fixtureRoot,
		bin:  filepath.Join(fixtureRoot, "bin", "camp-buzz"),
		env: []string{
			"HOME=" + filepath.Join(fixtureRoot, "home"),
			"PATH=" + filepath.Join(fixtureRoot, "bin") + string(os.PathListSeparator) + os.Getenv("PATH"),
			"CAMP_ROOT=" + filepath.Join(fixtureRoot, "campaign"),
			"CAMP_BUZZ_ROOT=" + filepath.Join(fixtureRoot, "home", ".obey", "plugins", "camp-buzz"),
			"BUZZ_CHANNEL=",
			"BUZZ_RELAY_URL=",
			"BUZZ_PRIVATE_KEY=demo-nsec-not-real",
			"BUZZ_FAKE_LOG=" + filepath.Join(fixtureRoot, "fake-buzz.log"),
		},
	}
}

func (fixture *cliFixture) run(extraEnv []string, args ...string) (string, error) {
	cmd := exec.Command(fixture.bin, args...)
	cmd.Env = append(append(append([]string{}, os.Environ()...), fixture.env...), extraEnv...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.String(), err
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		data, readErr := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readErr == nil && strings.Contains(string(data), "camp-buzz") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func mustRun(t *testing.T, fixture *cliFixture, extraEnv []string, args ...string) string {
	t.Helper()
	out, err := fixture.run(extraEnv, args...)
	if err != nil {
		t.Fatalf("camp-buzz %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func assertFailure(t *testing.T, output string, err error, wants ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected failure; output:\n%s", output)
	}
	assertOutput(t, output, wants...)
}

func assertOutput(t *testing.T, output string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
