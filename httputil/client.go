package httputil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultClientTimeout = 30 * time.Second

var defaultClient = &http.Client{
	Timeout: defaultClientTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func GetBytesFromUrl(url string) ([]byte, error) {
	return GetBytesFromUrlWithClient(defaultClient, url)
}

func GetBytesFromUrlWithTimeout(url string, timeout time.Duration) ([]byte, error) {
	return GetBytesFromUrlWithContextTimeout(context.Background(), url, timeout)
}

func GetBytesFromUrlWithClient(client *http.Client, url string) ([]byte, error) {
	return GetBytesFromUrlWithContext(context.Background(), client, url)
}

// GetBytesFromUrlWithContext 支持通过 context 取消的 HTTP GET 请求
func GetBytesFromUrlWithContext(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	if client == nil {
		client = defaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, url: %s, status code: %d", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetBytesFromUrlWithContextTimeout 同时支持 context 取消和超时
func GetBytesFromUrlWithContextTimeout(ctx context.Context, url string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = defaultClientTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return GetBytesFromUrlWithContext(ctx, defaultClient, url)
}

func GetStringFromUrl(url string) (string, error) {
	return GetStringFromUrlWithClient(defaultClient, url)
}

func GetStringFromUrlWithTimeout(url string, timeout time.Duration) (string, error) {
	return GetStringFromUrlWithContextTimeout(context.Background(), url, timeout)
}

func GetStringFromUrlWithClient(client *http.Client, url string) (string, error) {
	return GetStringFromUrlWithContext(context.Background(), client, url)
}

// GetStringFromUrlWithContext 支持通过 context 取消的 HTTP GET 请求，返回字符串
func GetStringFromUrlWithContext(ctx context.Context, client *http.Client, url string) (string, error) {
	data, err := GetBytesFromUrlWithContext(ctx, client, url)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetStringFromUrlWithContextTimeout 同时支持 context 取消和超时，返回字符串
func GetStringFromUrlWithContextTimeout(ctx context.Context, url string, timeout time.Duration) (string, error) {
	data, err := GetBytesFromUrlWithContextTimeout(ctx, url, timeout)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
