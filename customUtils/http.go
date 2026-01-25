package customUtils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/weiweimhy/go-utils/logger"
	"go.uber.org/zap"
)

var (
	// DefaultHttpClient 提供带超时的、生产环境安全的 HTTP 客户端。
	// 避免直接使用 http.Get/http.DefaultClient，因为它们没有默认超时。
	DefaultHttpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// GetBytesFromUrl 使用 Context 请求 URL 并返回字节流
func GetBytesFromUrl(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := DefaultHttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logger.L().Warn("failed to close http response body",
				zap.String("url", url),
				zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status: %d, url: %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

// GetStringFromUrl 使用 Context 请求 URL 并返回字符串
func GetStringFromUrl(ctx context.Context, url string) (string, error) {
	data, err := GetBytesFromUrl(ctx, url)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
