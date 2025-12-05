package runtime

import (
	"fmt"
	"runtime/debug"
)

func GetVersion() {
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Println(info.Main.Version)
	}
}

