package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Obedience-Corp/camp-buzz/internal/version"
)

func TestRootHelp(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected help output")
	}
}

func TestVersion(t *testing.T) {
	build := version.Current()
	want := fmt.Sprintf("camp-buzz %s (%s) built %s\n", build.Version, build.Commit, build.BuildDate)
	for _, args := range [][]string{{"version"}, {"--version"}} {
		cmd := newRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != want {
			t.Fatalf("%v output = %q, want %q", args, got, want)
		}
	}
}
