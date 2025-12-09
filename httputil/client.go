package httputil

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultClientTimeout = 30 * time.Second

func GetBytesFromUrl(url string) ([]byte, error) {
	return GetBytesFromUrlWithTimeout(url, defaultClientTimeout)
}

func GetBytesFromUrlWithTimeout(url string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = defaultClientTimeout
	}
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, url: %s, status code: %d", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func GetStringFromUrl(url string) (string, error) {
	return GetStringFromUrlWithTimeout(url, defaultClientTimeout)
}

func GetStringFromUrlWithTimeout(url string, timeout time.Duration) (string, error) {
	data, err := GetBytesFromUrlWithTimeout(url, timeout)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
