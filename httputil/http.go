package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var (
	// DefaultHTTPClient 提供带超时的、生产环境安全的 HTTP 客户端。
	// 避免直接使用 http.Get/http.DefaultClient，因为它们没有默认超时。
	DefaultHTTPClient = &http.Client{
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

	// ErrResponseTooLarge is returned when an Options.MaxBytes limit is exceeded.
	ErrResponseTooLarge = errors.New("httputil: response body exceeds max bytes")
)

// StatusError describes an HTTP response status outside the allowed set.
type StatusError struct {
	StatusCode int
	URL        string
	Body       []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("httputil: unexpected status %d for %s", e.StatusCode, e.URL)
}

// Options controls HTTP helper behavior.
type Options struct {
	Client        *http.Client
	MaxBytes      int64
	Headers       http.Header
	AllowedStatus []int
}

func (opts Options) client() *http.Client {
	if opts.Client != nil {
		return opts.Client
	}
	return DefaultHTTPClient
}

func (opts Options) statusAllowed(status int) bool {
	if len(opts.AllowedStatus) == 0 {
		return status == http.StatusOK
	}
	for _, allowed := range opts.AllowedStatus {
		if status == allowed {
			return true
		}
	}
	return false
}

// GetBytesFromURL 使用 Context 请求 URL 并返回字节流。
func GetBytesFromURL(ctx context.Context, url string) ([]byte, error) {
	return GetBytes(ctx, url, Options{})
}

// GetBytes requests url with GET and returns the response body.
func GetBytes(ctx context.Context, url string, opts Options) ([]byte, error) {
	return DoBytes(ctx, http.MethodGet, url, nil, opts)
}

// DoBytes executes an HTTP request and returns the response body.
func DoBytes(ctx context.Context, method, url string, body io.Reader, opts Options) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	for key, values := range opts.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := opts.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := readBody(resp.Body, opts.MaxBytes)
	if err != nil {
		return nil, err
	}

	if !opts.statusAllowed(resp.StatusCode) {
		return nil, &StatusError{StatusCode: resp.StatusCode, URL: url, Body: data}
	}

	return data, nil
}

// GetStringFromURL 使用 Context 请求 URL 并返回字符串。
func GetStringFromURL(ctx context.Context, url string) (string, error) {
	return GetString(ctx, url, Options{})
}

// GetString requests url with GET and returns the response body as a string.
func GetString(ctx context.Context, url string, opts Options) (string, error) {
	data, err := GetBytes(ctx, url, opts)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DoJSON sends an optional JSON request body and decodes a JSON response.
func DoJSON[T any](ctx context.Context, method, url string, request any, opts Options) (T, error) {
	var zero T
	var body io.Reader
	if request != nil {
		data, err := json.Marshal(request)
		if err != nil {
			return zero, err
		}
		body = bytes.NewReader(data)
		if opts.Headers == nil {
			opts.Headers = make(http.Header)
		}
		if opts.Headers.Get("Content-Type") == "" {
			opts.Headers.Set("Content-Type", "application/json")
		}
	}
	if opts.Headers == nil {
		opts.Headers = make(http.Header)
	}
	if opts.Headers.Get("Accept") == "" {
		opts.Headers.Set("Accept", "application/json")
	}

	data, err := DoBytes(ctx, method, url, body, opts)
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

// GetJSON requests url with GET and decodes the JSON response.
func GetJSON[T any](ctx context.Context, url string, opts Options) (T, error) {
	return DoJSON[T](ctx, http.MethodGet, url, nil, opts)
}

// PostJSON posts value as JSON and decodes the JSON response.
func PostJSON[T any](ctx context.Context, url string, value any, opts Options) (T, error) {
	return DoJSON[T](ctx, http.MethodPost, url, value, opts)
}

func readBody(body io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(body)
	}
	limited := io.LimitReader(body, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrResponseTooLarge
	}
	return data, nil
}
