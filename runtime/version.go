package runtime

import (
	"runtime/debug"
)

func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

