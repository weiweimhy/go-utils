package httputil

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/weiweimhy/go-utils/v2/filesystem"
	"github.com/weiweimhy/go-utils/v2/logger"
	"go.uber.org/zap"
)

const (
	defaultDownloadTimeout = 60 * time.Second
)

func DownloadFile(url string, path string) error {
	return DownloadFileWithTimeout(url, path, defaultDownloadTimeout)
}

func DownloadFileWithTimeout(url string, path string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultDownloadTimeout
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return downloadFile(client, url, path)
}

// downloadFile 内部下载函数，供 DownloadFile 和 DownloadTask 使用
func downloadFile(client *http.Client, url string, path string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.L().Warn("failed to close http response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	err = filesystem.CreateDir(path)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.L().Warn("failed to close file", zap.String("path", path), zap.Error(err))
		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	return err
}
