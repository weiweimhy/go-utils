package httputil

import (
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
	if timeout <= 0 {
		timeout = defaultClientTimeout
	}
	client := &http.Client{Timeout: timeout}
	return GetBytesFromUrlWithClient(client, url)
}

func GetBytesFromUrlWithClient(client *http.Client, url string) ([]byte, error) {
	if client == nil {
		client = defaultClient
	}

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
	return GetStringFromUrlWithClient(defaultClient, url)
}

func GetStringFromUrlWithTimeout(url string, timeout time.Duration) (string, error) {
	data, err := GetBytesFromUrlWithTimeout(url, timeout)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func GetStringFromUrlWithClient(client *http.Client, url string) (string, error) {
	data, err := GetBytesFromUrlWithClient(client, url)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
