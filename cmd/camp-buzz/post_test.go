package main

import (
	"strings"
	"testing"

	"github.com/Obedience-Corp/camp-buzz/internal/buzzcli"
	"github.com/Obedience-Corp/camp-buzz/internal/config"
)

func TestReadPostBodyBoundaries(t *testing.T) {
	atLimit := strings.Repeat("x", buzzcli.MaxContentBytes)
	tests := []struct {
		name, message, stdin, want string
		fromHook                   bool
		wantErr                    bool
	}{
		{name: "message at limit", message: atLimit, want: atLimit},
		{name: "message over limit", message: atLimit + "x", wantErr: true},
		{name: "stdin at limit", stdin: atLimit, want: atLimit},
		{name: "stdin over limit", stdin: atLimit + "x", wantErr: true},
		{name: "empty direct", stdin: " \n\t", wantErr: true},
		{name: "empty hook default", fromHook: true, want: "Festival status update"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPostBody(tt.message, strings.NewReader(tt.stdin), tt.fromHook)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("body length = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestAppendFooterDefaultsForHook(t *testing.T) {
	body, err := (&postOptions{}).appendFooter("Festival status update", config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"festival: -", "task: -", "path: -", "gate: n/a"} {
		if !strings.Contains(body, want) {
			t.Fatalf("footer missing %q: %s", want, body)
		}
	}
}

func TestValidateFooterRejectsMalformedFields(t *testing.T) {
	tests := []struct {
		name, festival, task, path, gate string
	}{
		{name: "festival newline", festival: "FE1\ninjected", task: "-", path: "-", gate: "n/a"},
		{name: "task newline", festival: "FE1", task: "T1\ninjected", path: "-", gate: "pass"},
		{name: "absolute path", festival: "FE1", task: "T1", path: "/private/work", gate: "pass"},
		{name: "escaping path", festival: "FE1", task: "T1", path: "../other", gate: "pass"},
		{name: "invalid gate", festival: "FE1", task: "T1", path: "festivals/active/FE1", gate: "green"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFooter(tt.festival, tt.task, tt.path, tt.gate); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestShowDisplaySanitizesInvalidValues(t *testing.T) {
	if got := safeDisplay("FE1\n\x1b[31m", "fallback"); got != `FE1\n\x1b[31m` {
		t.Fatalf("safeDisplay = %q", got)
	}
	if got := displayRelay("https://user:secret@example.com"); strings.Contains(got, "secret") || !strings.Contains(got, "invalid") {
		t.Fatalf("displayRelay = %q", got)
	}
	if got := displayChannel("not-a-uuid"); !strings.Contains(got, "invalid") {
		t.Fatalf("displayChannel = %q", got)
	}
}
