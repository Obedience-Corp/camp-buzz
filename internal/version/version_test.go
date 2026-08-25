package version

import "testing"

func TestResolveModuleVersion(t *testing.T) {
	tests := []struct {
		name          string
		linkedVersion string
		moduleVersion string
		want          string
	}{
		{name: "linked release wins", linkedVersion: "1.2.3", moduleVersion: "v9.9.9", want: "1.2.3"},
		{name: "go install module", linkedVersion: "dev", moduleVersion: "v0.1.1", want: "0.1.1"},
		{name: "pseudo version", linkedVersion: "dev", moduleVersion: "v0.0.0-20260825220534-ed679b907b7d", want: "0.0.0-20260825220534-ed679b907b7d"},
		{name: "local build", linkedVersion: "dev", moduleVersion: "(devel)", want: "dev"},
		{name: "missing build info", linkedVersion: "dev", moduleVersion: "", want: "dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			build := resolveModuleVersion(Build{
				Version:   test.linkedVersion,
				Commit:    "commit",
				BuildDate: "date",
			}, test.moduleVersion)
			if build.Version != test.want {
				t.Fatalf("Version = %q, want %q", build.Version, test.want)
			}
			if build.Commit != "commit" || build.BuildDate != "date" {
				t.Fatalf("unrelated metadata changed: %#v", build)
			}
		})
	}
}
