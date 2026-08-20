package logger

import (
	"runtime"
	"strconv"
	"strings"
)

// GoroutineID returns the current goroutine ID for debugging and logging only.
// Do not rely on goroutine IDs in business logic.
func GoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields := strings.Fields(string(buf[:n]))
	if len(fields) < 2 || fields[0] != "goroutine" {
		return 0
	}
	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}
