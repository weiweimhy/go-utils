package logger

import (
	"runtime"
	"strconv"
	"strings"
)

// GetGoroutineID 获取当前 goroutine ID，用于调试日志
// 注意：此方法仅用于调试和日志记录，不要在业务逻辑中依赖 goroutine ID
func GetGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// 格式: "goroutine 123 [running]:\n..."
	s := string(buf[:n])
	s = strings.TrimPrefix(s, "goroutine ")
	s = s[:strings.IndexByte(s, ' ')]
	id, _ := strconv.ParseUint(s, 10, 64)
	return id
}
