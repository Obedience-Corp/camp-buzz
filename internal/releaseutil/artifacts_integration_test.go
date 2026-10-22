//go:build integration

package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyArtifactsRequiresCompleteValidMatrix(t *testing.T) {
	dir := t.TempDir()
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var checksums strings.Builder
	for target, name := range expectedArchiveNames("0.1.0") {
		binary := buildTargetBinary(t, repoRoot, dir, target)
		archive := filepath.Join(dir, name)
		writeTestArchive(t, archive, binary)
		data, err := os.ReadFile(archive)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&checksums, "%x  %s\n", sha256.Sum256(data), name)
	}
	if err := os.WriteFile(filepath.Join(dir, "checksums.txt"), []byte(checksums.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyArtifacts(dir, "v0.1.0"); err != nil {
		t.Fatalf("complete matrix failed: %v", err)
	}
	missing := expectedArchiveNames("0.1.0")[releaseTarget{OS: "macOS", Arch: "x86_64"}]
	if err := os.Remove(filepath.Join(dir, missing)); err != nil {
		t.Fatal(err)
	}
	if err := verifyArtifacts(dir, "v0.1.0"); err == nil {
		t.Fatal("incomplete snapshot passed artifact validation")
	}
}

func buildTargetBinary(t *testing.T, repoRoot, dir string, target releaseTarget) []byte {
	t.Helper()
	goos := target.OS
	goarch := target.Arch
	if goos == "macOS" {
		goos = "darwin"
	}
	if goarch == "x86_64" {
		goarch = "amd64"
	}
	path := filepath.Join(dir, "camp-buzz-"+goos+"-"+goarch)
	cmd := exec.Command("go", "build", "-o", path, "./cmd/camp-buzz")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s/%s: %v\n%s", goos, goarch, err, output)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeTestArchive(t *testing.T, path string, binary []byte) {
	t.Helper()
	writeNamedTestArchive(t, path, binary, "camp-buzz")
}

func writeNamedTestArchive(t *testing.T, path string, binary []byte, binaryName string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gz)
	for _, name := range requiredPackageFiles {
		content := []byte("fixture")
		mode := int64(0o644)
		if name == "camp-buzz" {
			name = binaryName
			content = binary
			mode = 0o755
		}
		header := &tar.Header{Name: name, Mode: mode, Size: int64(len(content))}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsCandidateArchiveRequiresPEBinary(t *testing.T) {
	dir := t.TempDir()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	target := releaseTarget{OS: "windows", Arch: "x86_64"}
	binary := buildTargetBinary(t, root, dir, target)
	archive := filepath.Join(dir, "windows.tar.gz")
	writeNamedTestArchive(t, archive, binary, "camp-buzz.exe")
	if err := verifyArchive(archive, target); err != nil {
		t.Fatalf("valid Windows payload: %v", err)
	}
	corrupt := append([]byte(nil), binary...)
	corrupt[0] ^= 1
	writeNamedTestArchive(t, archive, corrupt, "camp-buzz.exe")
	if err := verifyArchive(archive, target); err == nil {
		t.Fatal("corrupt PE passed")
	}
	writeNamedTestArchive(t, archive, binary, "camp-buzz")
	if err := verifyArchive(archive, target); err == nil {
		t.Fatal("Windows archive without exe name passed")
	}
}
