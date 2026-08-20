//go:build !windows

package pathutil

func isReparsePoint(string) bool {
	return false
}
