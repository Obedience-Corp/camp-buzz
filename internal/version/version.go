package version

import (
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

type Build struct {
	Version   string
	Commit    string
	BuildDate string
}

func Current() Build {
	build := Build{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return build
	}
	return resolveModuleVersion(build, info.Main.Version)
}

func resolveModuleVersion(build Build, moduleVersion string) Build {
	if build.Version != "dev" || moduleVersion == "" || moduleVersion == "(devel)" {
		return build
	}
	build.Version = strings.TrimPrefix(moduleVersion, "v")
	return build
}
