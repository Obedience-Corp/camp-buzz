package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/elf"
	"debug/macho"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const maxPackagedBinaryBytes = 64 << 20

type releaseTarget struct {
	OS   string
	Arch string
}

var releaseTargets = []releaseTarget{
	{OS: "linux", Arch: "x86_64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "macOS", Arch: "x86_64"},
	{OS: "macOS", Arch: "arm64"},
}

var requiredPackageFiles = []string{
	"camp-buzz",
	"README.md",
	"LICENSE",
	"NOTICE",
	"assets/templates/fest-hooks.example.yaml",
	"assets/templates/status-footer.example.md",
	"completions/_camp-buzz",
	"completions/camp-buzz.bash",
	"completions/camp-buzz.fish",
}

func verifyArtifacts(dir, tag string) error {
	if err := validateExplicitTag(tag); err != nil {
		return fmt.Errorf("artifact tag: %w", err)
	}
	version := strings.TrimPrefix(tag, "v")
	expected := expectedArchiveNames(version)
	checksums, err := readChecksums(filepath.Join(dir, "checksums.txt"))
	if err != nil {
		return err
	}
	for target, name := range expected {
		path := filepath.Join(dir, name)
		if err := verifyChecksum(path, name, checksums); err != nil {
			return err
		}
		if err := verifyArchive(path, target); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	if len(checksums) != len(expected) {
		return fmt.Errorf("checksums.txt contains %d entries, want %d", len(checksums), len(expected))
	}
	fmt.Printf("Verified %d release archives and checksums for %s\n", len(expected), tag)
	return nil
}

func expectedArchiveNames(version string) map[releaseTarget]string {
	names := make(map[releaseTarget]string, len(releaseTargets))
	for _, target := range releaseTargets {
		names[target] = fmt.Sprintf("camp-buzz-%s-%s-%s.tar.gz", version, target.OS, target.Arch)
	}
	return names
}

func readChecksums(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	checksums := make(map[string]string)
	for lineNumber, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("checksums.txt line %d is malformed", lineNumber+1)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil || len(fields[0]) != sha256.Size*2 {
			return nil, fmt.Errorf("checksums.txt line %d has invalid SHA-256", lineNumber+1)
		}
		if _, exists := checksums[fields[1]]; exists {
			return nil, fmt.Errorf("checksums.txt contains duplicate %s", fields[1])
		}
		checksums[fields[1]] = fields[0]
	}
	return checksums, nil
}

func verifyChecksum(path, name string, checksums map[string]string) error {
	want, ok := checksums[name]
	if !ok {
		return fmt.Errorf("checksums.txt is missing %s", name)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open expected archive %s: %w", name, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash %s: %w", name, err)
	}
	if got := hex.EncodeToString(hash.Sum(nil)); got != want {
		return fmt.Errorf("checksum mismatch for %s", name)
	}
	return nil
}

func verifyArchive(path string, target releaseTarget) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gz.Close()
	files, binary, err := readArchive(tar.NewReader(gz))
	if err != nil {
		return err
	}
	for _, required := range requiredPackageFiles {
		if !files[required] {
			return fmt.Errorf("missing packaged file %s", required)
		}
	}
	if len(files) != len(requiredPackageFiles) {
		return fmt.Errorf("archive contains %d files, want exactly %d", len(files), len(requiredPackageFiles))
	}
	return verifyBinaryTarget(binary, target)
}

func readArchive(reader *tar.Reader) (map[string]bool, []byte, error) {
	files := make(map[string]bool)
	var binary []byte
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("read archive: %w", err)
		}
		name := filepath.ToSlash(header.Name)
		if name == "" || path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") {
			return nil, nil, fmt.Errorf("unsafe archive path %q", name)
		}
		if !header.FileInfo().Mode().IsRegular() {
			return nil, nil, fmt.Errorf("archive entry %s is not a regular file", name)
		}
		if files[name] {
			return nil, nil, fmt.Errorf("archive contains duplicate file %s", name)
		}
		files[name] = true
		if name == "camp-buzz" {
			if header.FileInfo().Mode().Perm()&0o111 == 0 {
				return nil, nil, fmt.Errorf("packaged binary is not executable")
			}
			binary, err = io.ReadAll(io.LimitReader(reader, maxPackagedBinaryBytes+1))
			if err != nil {
				return nil, nil, fmt.Errorf("read packaged binary: %w", err)
			}
			if len(binary) > maxPackagedBinaryBytes {
				return nil, nil, fmt.Errorf("packaged binary exceeds %d bytes", maxPackagedBinaryBytes)
			}
		}
	}
	return files, binary, nil
}

func verifyBinaryTarget(binary []byte, target releaseTarget) error {
	if len(binary) == 0 {
		return fmt.Errorf("packaged binary is empty")
	}
	if target.OS == "linux" {
		file, err := elf.NewFile(bytes.NewReader(binary))
		if err != nil {
			return fmt.Errorf("parse linux binary: %w", err)
		}
		defer file.Close()
		want := elf.EM_X86_64
		if target.Arch == "arm64" {
			want = elf.EM_AARCH64
		}
		if file.Machine != want {
			return fmt.Errorf("linux binary architecture is %s, want %s", file.Machine, want)
		}
		return nil
	}
	file, err := macho.NewFile(bytes.NewReader(binary))
	if err != nil {
		return fmt.Errorf("parse macOS binary: %w", err)
	}
	defer file.Close()
	want := macho.CpuAmd64
	if target.Arch == "arm64" {
		want = macho.CpuArm64
	}
	if file.Cpu != want {
		return fmt.Errorf("macOS binary architecture is %s, want %s", file.Cpu, want)
	}
	return nil
}
