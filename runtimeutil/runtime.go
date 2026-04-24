package runtimeutil

import (
	"runtime/debug"
)

// GetVersion returns the main module version from build info.
// It returns an empty string when build metadata is unavailable.
func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}
