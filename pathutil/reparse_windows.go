//go:build windows

package pathutil

import "syscall"

const fileAttributeReparsePoint = 0x0400

func isReparsePoint(path string) bool {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := syscall.GetFileAttributes(pathPtr)
	return err == nil && attrs&fileAttributeReparsePoint != 0
}
