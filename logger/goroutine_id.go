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
	// 格式: "goroutine 123 [running]:\n..."
	s := string(buf[:n])
	s = strings.TrimPrefix(s, "goroutine ")
	s = s[:strings.IndexByte(s, ' ')]
	id, _ := strconv.ParseUint(s, 10, 64)
	return id
}
