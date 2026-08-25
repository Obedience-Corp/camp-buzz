//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseGitPreconditions(t *testing.T) {
	repo := newReleaseRepository(t)
	t.Chdir(repo)
	if err := ensureCleanTree(); err != nil {
		t.Fatalf("clean tree: %v", err)
	}
	if err := ensureBranch("main"); err != nil {
		t.Fatalf("main branch: %v", err)
	}
	if err := fetchOriginRefs(); err != nil {
		t.Fatalf("fetch origin refs: %v", err)
	}
	if err := ensureSyncedWithOriginMain(); err != nil {
		t.Fatalf("synchronized main: %v", err)
	}
	if err := ensureTagAbsent("v0.1.0"); err != nil {
		t.Fatalf("absent tag: %v", err)
	}

	writeTestFile(t, filepath.Join(repo, "dirty.txt"), "dirty")
	if err := ensureCleanTree(); err == nil {
		t.Fatal("dirty tree passed release precondition")
	}
	if err := os.Remove(filepath.Join(repo, "dirty.txt")); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, repo, "checkout", "-b", "feature")
	if err := ensureBranch("main"); err == nil || !strings.Contains(err.Error(), "current: feature") {
		t.Fatalf("branch error = %v", err)
	}
	runGitTest(t, repo, "checkout", "main")
	writeTestFile(t, filepath.Join(repo, "ahead.txt"), "ahead")
	runGitTest(t, repo, "add", "ahead.txt")
	runGitTest(t, repo, "commit", "-m", "ahead")
	if err := ensureSyncedWithOriginMain(); err == nil {
		t.Fatal("ahead main passed synchronization precondition")
	}
	runGitTest(t, repo, "tag", "v0.1.0")
	if err := ensureTagAbsent("v0.1.0"); err == nil {
		t.Fatal("existing local tag passed absence precondition")
	}
}

func newReleaseRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	repo := filepath.Join(root, "work")
	runGitTest(t, root, "init", "--bare", remote)
	runGitTest(t, root, "init", "-b", "main", repo)
	runGitTest(t, repo, "config", "user.name", "Release Test")
	runGitTest(t, repo, "config", "user.email", "release-test@example.invalid")
	writeTestFile(t, filepath.Join(repo, "README.md"), "release fixture")
	runGitTest(t, repo, "add", "README.md")
	runGitTest(t, repo, "commit", "-m", "initial")
	runGitTest(t, repo, "remote", "add", "origin", remote)
	runGitTest(t, repo, "push", "-u", "origin", "main")
	return repo
}

func runGitTest(t *testing.T, directory string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
