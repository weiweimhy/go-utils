package httputil

import (
	"context"
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
	return DownloadFileWithContextTimeout(context.Background(), url, path, timeout)
}

func DownloadFileWithClient(client *http.Client, url string, path string) error {
	return DownloadFileWithContext(context.Background(), client, url, path)
}

// DownloadFileWithContext 支持通过 context 取消的下载
func DownloadFileWithContext(ctx context.Context, client *http.Client, url string, path string) error {
	if client == nil {
		client = defaultDownloadClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	err = filesystem.CreateParentDir(path)
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

// DownloadFileWithContextTimeout 同时支持 context 取消和超时
func DownloadFileWithContextTimeout(ctx context.Context, url string, path string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultDownloadTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return DownloadFileWithContext(ctx, defaultDownloadClient, url, path)
}
