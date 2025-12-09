package httputil

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/weiweimhy/go-utils/v2/filesystem"
)

const defaultDownloadTimeout = 60 * time.Second

var defaultDownloadClient = &http.Client{
	Timeout: defaultDownloadTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func DownloadFile(url string, path string) error {
	return DownloadFileWithClient(defaultDownloadClient, url, path)
}

func DownloadFileWithTimeout(url string, path string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultDownloadTimeout
	}
	client := &http.Client{Timeout: timeout}
	return DownloadFileWithClient(client, url, path)
}

func DownloadFileWithClient(client *http.Client, url string, path string) error {
	if client == nil {
		client = defaultDownloadClient
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
