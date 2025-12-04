package customUtils

import (
	"os"
	"path/filepath"
	"strings"
)

func IsFileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SaveToFile(path string, data []byte) error {
	err := CreateDir(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0666)
}

func GetFileBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return GetBase64FromBytes(data), nil
}

func GetStringFormFile(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetNameWithPathWithoutExt
//   - 包含完整路径
//   - 不包含文件后缀
//   - 示例："/home/user/docs/report.pdf" → "/home/user/docs/report"
func GetNameWithPathWithoutExt(fullPath string) string {
	// 1. 取目录 + 文件名（不含后缀）
	dir := filepath.Dir(fullPath)
	base := filepath.Base(fullPath)

	// 2. 去掉后缀
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

	// 3. 拼接：目录 + 文件名（无后缀）
	if dir == "." {
		return nameWithoutExt
	}
	return filepath.Join(dir, nameWithoutExt)
}
