// Package version holds the CLI's semver string.
//
// Set via ldflags at build time:
//
//	go build -ldflags "-X github.com/aiaikit/speckit/internal/version.Version=1.2.3"
//
// When ldflags is not set, falls back to runtime/debug.ReadBuildInfo.
// If neither yields a version, the string is "dev".
package version

import "runtime/debug"

// Version is the CLI's semver string. Overwritten by -ldflags at build
// time, otherwise resolved from runtime/debug.ReadBuildInfo, otherwise "dev".
var Version = resolveDefault()

func resolveDefault() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	v := info.Main.Version
	if v == "" || v == "(devel)" {
		return "dev"
	}
	return v
}

// Set overwrites Version. Called from main.go when a build flag is parsed
// or when the binary needs to record a self-upgrade target version.
func Set(v string) {
	Version = v
}