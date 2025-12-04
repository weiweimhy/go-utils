package logger

import (
	"runtime"
	"strconv"
	"strings"
)

func GetGoroutineID() uint64 {
	buf := make([]byte, 64)
	runtime.Stack(buf, false)
	// goroutine 123 [running]:
	fields := strings.Fields(string(buf))
	if len(fields) >= 2 {
		id, _ := strconv.ParseUint(fields[1], 10, 64)
		return id
	}
	return 0
}
