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

	"github.com/weiweimhy/go-utils/v5/securityutil"
	"github.com/weiweimhy/go-utils/v5/streamutil"
)

const (
	// DefaultTimeout is used by NewClient when no positive timeout is supplied.
	DefaultTimeout time.Duration = 30 * time.Second

	defaultErrorBodyMaxBytes int64 = 64 << 10
)

var (
	// DefaultHTTPClient 提供带超时的、生产环境安全的 HTTP 客户端。
	// 避免直接使用 http.Get/http.DefaultClient，因为它们没有默认超时。
	DefaultHTTPClient = &http.Client{
		Timeout:   DefaultTimeout,
		Transport: newDefaultTransport(),
	}

	// ErrResponseTooLarge is returned when an Options.MaxBytes limit is exceeded.
	ErrResponseTooLarge = errors.New("httputil: response body exceeds max bytes")
)

func newDefaultTransport() *http.Transport {
	return &http.Transport{
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
	}
}

// NewClient creates an independent client with the package's safe transport
// defaults. A non-positive timeout uses DefaultTimeout.
func NewClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &http.Client{Timeout: timeout, Transport: newDefaultTransport()}
}

// StatusError describes an HTTP response status outside the allowed set.
type StatusError struct {
	StatusCode int
	// URL is redacted before the error is returned.
	URL string
	// Body is available only when Options.CaptureErrorBody is true.
	Body []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("httputil: unexpected status %d for %s", e.StatusCode, e.URL)
}

// Options controls HTTP helper behavior.
type Options struct {
	Client            *http.Client
	MaxBytes          int64
	Headers           http.Header
	AllowedStatus     []int
	SuccessStatus     func(status int) bool
	CaptureErrorBody  bool
	MaxErrorBodyBytes int64
	RedactURL         func(rawURL string) string
}

func (opts Options) client() *http.Client {
	if opts.Client != nil {
		return opts.Client
	}
	return DefaultHTTPClient
}

func (opts Options) statusAllowed(status int) bool {
	if opts.SuccessStatus != nil {
		return opts.SuccessStatus(status)
	}
	if len(opts.AllowedStatus) == 0 {
		return status >= http.StatusOK && status < http.StatusMultipleChoices
	}
	for _, allowed := range opts.AllowedStatus {
		if status == allowed {
			return true
		}
	}
	return false
}

func (opts Options) redactedURL(rawURL string) string {
	if opts.RedactURL != nil {
		return opts.RedactURL(rawURL)
	}
	return securityutil.RedactURLQuery(rawURL)
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

	if !opts.statusAllowed(resp.StatusCode) {
		var data []byte
		if opts.CaptureErrorBody {
			maxBytes := opts.MaxErrorBodyBytes
			if maxBytes <= 0 {
				maxBytes = defaultErrorBodyMaxBytes
			}
			if opts.MaxBytes > 0 && opts.MaxBytes < maxBytes {
				maxBytes = opts.MaxBytes
			}
			data, err = readBody(resp.Body, maxBytes)
			if err != nil && !errors.Is(err, ErrResponseTooLarge) {
				return nil, err
			}
		}
		return nil, &StatusError{StatusCode: resp.StatusCode, URL: opts.redactedURL(url), Body: data}
	}

	data, err := readBody(resp.Body, opts.MaxBytes)
	if err != nil {
		return nil, err
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
	if opts.Headers != nil {
		opts.Headers = opts.Headers.Clone()
	} else {
		opts.Headers = make(http.Header)
	}
	if request != nil {
		data, err := json.Marshal(request)
		if err != nil {
			return zero, err
		}
		body = bytes.NewReader(data)
		if opts.Headers.Get("Content-Type") == "" {
			opts.Headers.Set("Content-Type", "application/json")
		}
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
	data, err := streamutil.ReadAllLimit(body, maxBytes)
	if errors.Is(err, streamutil.ErrLimitExceeded) {
		return nil, ErrResponseTooLarge
	}
	return data, err
}
