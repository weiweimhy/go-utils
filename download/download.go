package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/weiweimhy/go-utils/v6/fsutil"
	"github.com/weiweimhy/go-utils/v6/httputil"
	"github.com/weiweimhy/go-utils/v6/securityutil"
	"github.com/weiweimhy/go-utils/v6/task"
)

var (
	// ErrManagerClosed is returned when adding work to a closed manager.
	ErrManagerClosed = errors.New("download: manager is closed")
	// ErrManagerNotStarted is returned when adding work before Start.
	ErrManagerNotStarted = errors.New("download: manager not started")
	// ErrResponseTooLarge is returned when a response exceeds the configured download limit.
	ErrResponseTooLarge = errors.New("download: response body exceeds max bytes")
	// ErrCloseTimeout is returned when a cancelled download task does not finish in time.
	ErrCloseTimeout = errors.New("download: manager close timeout")
)

// DefaultMaxBytes is the maximum response size accepted by a manager unless
// WithMaxBytes overrides it.
const DefaultMaxBytes int64 = 1 << 30

// DownloadTask 表示一个下载任务
type DownloadTask struct {
	URL      string
	SavePath string
	Callback func(url, savePath string, err error)
}

// DownloadManager 管理大规模并发下载任务
type DownloadManager struct {
	mu       sync.Mutex
	pool     *task.WorkerPool
	client   *http.Client
	delay    time.Duration
	maxBytes int64

	workers    int
	bufferSize int
	log        *slog.Logger

	wg          sync.WaitGroup
	adders      sync.WaitGroup
	closeOnce   sync.Once
	closeDone   chan struct{}
	closeErr    error
	closed      bool
	callbackIDs sync.Map

	ctx    context.Context
	cancel context.CancelFunc
}

// DMOption 定义 DownloadManager 的配置选项
type DMOption func(*DownloadManager)

// WithWorkers 设置工作协程数量
func WithWorkers(w int) DMOption {
	return func(dm *DownloadManager) {
		if w > 0 {
			dm.workers = w
		}
	}
}

// WithDelay 设置每个下载任务之间的延迟
func WithDelay(d time.Duration) DMOption {
	return func(dm *DownloadManager) { dm.delay = d }
}

// WithClient 设置自定义 HTTP 客户端
func WithClient(c *http.Client) DMOption {
	return func(dm *DownloadManager) {
		if c != nil {
			dm.client = c
		}
	}
}

// WithMaxBytes sets the maximum bytes written for each download. A negative
// value explicitly disables the limit.
func WithMaxBytes(maxBytes int64) DMOption {
	return func(dm *DownloadManager) {
		if maxBytes != 0 {
			dm.maxBytes = maxBytes
		}
	}
}

// WithLogger sets an optional logger for manager lifecycle and failures.
// When unset, the manager stays silent by default.
func WithLogger(log *slog.Logger) DMOption {
	return func(dm *DownloadManager) { dm.log = log }
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(opts ...DMOption) *DownloadManager {
	dm := &DownloadManager{
		client:     httputil.DefaultHTTPClient,
		workers:    20,
		bufferSize: 100,
		maxBytes:   DefaultMaxBytes,
		closeDone:  make(chan struct{}),
	}
	for _, opt := range opts {
		opt(dm)
	}
	return dm
}

// Start 启动下载管理器
func (dm *DownloadManager) Start(ctx context.Context) error {
	return dm.StartWithConfig(ctx, dm.workers, dm.bufferSize)
}

// StartWithConfig 使用指定配置启动下载管理器
func (dm *DownloadManager) StartWithConfig(ctx context.Context, workers, bufferSize int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.closed {
		return ErrManagerClosed
	}
	if dm.pool != nil {
		return nil // 已启动
	}
	if ctx == nil {
		ctx = context.Background()
	}

	dm.ctx, dm.cancel = context.WithCancel(ctx)
	dm.pool = task.NewWorkerPool(
		dm.ctx,
		task.WithWorkerCount(workers),
		task.WithBufferSize(bufferSize),
		task.WithName("download-manager"),
		task.WithLogger(dm.log),
	)

	if dm.log != nil {
		dm.log.InfoContext(dm.ctx, "download manager started", "workers", workers, "buffer", bufferSize)
	}
	return nil
}

// Add 添加下载任务
func (dm *DownloadManager) Add(url, savePath string) error {
	return dm.AddWithCallback(url, savePath, nil)
}

// AddWithCallback 添加带回调的下载任务
func (dm *DownloadManager) AddWithCallback(url, savePath string, callback func(url, savePath string, err error)) error {
	dm.mu.Lock()
	if dm.closed {
		dm.mu.Unlock()
		return ErrManagerClosed
	}
	if dm.pool == nil {
		dm.mu.Unlock()
		return ErrManagerNotStarted
	}
	dm.adders.Add(1)
	pool := dm.pool
	client := dm.client
	delay := dm.delay
	maxBytes := dm.maxBytes
	log := dm.log
	ctx := dm.ctx
	dm.mu.Unlock()
	defer dm.adders.Done()

	dm.wg.Add(1)

	submitted := pool.SubmitFunc(func(ctx context.Context) {
		defer dm.wg.Done()

		err := downloadFile(ctx, client, url, savePath, maxBytes)
		if err != nil && log != nil {
			log.WarnContext(ctx, "download failed", "url", securityutil.RedactURL(url), "path", savePath, "error", err)
		}

		if callback != nil {
			dm.invokeCallback(callback, url, savePath, err)
		}

		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
			}
		}
	})

	if !submitted {
		dm.wg.Done()
		if err := ctx.Err(); err != nil {
			return err
		}
		return ErrManagerClosed
	}

	return nil
}

// Wait 等待所有下载任务完成
func (dm *DownloadManager) Wait() {
	dm.wg.Wait()
}

// Close stops accepting work, cancels active downloads, and waits for their
// callbacks to finish. It returns ErrCloseTimeout if a task ignores cancellation.
// When called by a task callback, it starts shutdown and returns immediately to
// avoid waiting for that callback to return.
func (dm *DownloadManager) Close() error {
	dm.closeOnce.Do(dm.beginClose)
	if dm.inCallback() {
		return nil
	}

	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case <-dm.closeDone:
		dm.mu.Lock()
		defer dm.mu.Unlock()
		return dm.closeErr
	case <-timer.C:
		dm.mu.Lock()
		if dm.closeErr == nil {
			dm.closeErr = ErrCloseTimeout
		}
		dm.mu.Unlock()
		return ErrCloseTimeout
	}
}

func (dm *DownloadManager) invokeCallback(callback func(url, savePath string, err error), url, savePath string, err error) {
	callbackID := currentGoroutineID()
	if callbackID != 0 {
		dm.callbackIDs.Store(callbackID, struct{}{})
		defer dm.callbackIDs.Delete(callbackID)
	}
	callback(url, savePath, err)
}

func (dm *DownloadManager) inCallback() bool {
	callbackID := currentGoroutineID()
	if callbackID == 0 {
		return false
	}
	_, ok := dm.callbackIDs.Load(callbackID)
	return ok
}

func currentGoroutineID() uint64 {
	var buffer [64]byte
	n := runtime.Stack(buffer[:], false)
	fields := strings.Fields(string(buffer[:n]))
	if len(fields) < 2 || fields[0] != "goroutine" {
		return 0
	}
	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func (dm *DownloadManager) beginClose() {
	dm.mu.Lock()
	dm.closed = true
	pool := dm.pool
	cancel := dm.cancel
	dm.mu.Unlock()

	if pool != nil {
		// A zero timeout initiates pool shutdown and cancels task contexts without
		// waiting under the manager lifecycle lock.
		pool.Close(0)
	}
	if cancel != nil {
		cancel()
	}

	go func() {
		dm.adders.Wait()
		dm.wg.Wait()
		if pool != nil {
			// Wait for workers to finish queued cancellation callbacks. The return
			// value is false because shutdown was intentionally initiated above.
			pool.Close(30 * time.Second)
		}
		dm.mu.Lock()
		close(dm.closeDone)
		dm.mu.Unlock()
	}()
}

// downloadFile 下载单个文件
func downloadFile(ctx context.Context, client *http.Client, url string, path string, maxBytes int64) error {
	if client == nil {
		return fmt.Errorf("download: HTTP client is required")
	}
	redactedURL := securityutil.RedactURL(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &RequestError{URL: redactedURL, Err: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return &RequestError{URL: redactedURL, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: status %d", resp.StatusCode)
	}

	if err := writeResponseAtomic(path, resp.Body, maxBytes); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}
	return nil
}

// RequestError reports a download request failure without exposing credentials
// embedded in the URL.
type RequestError struct {
	URL string
	Err error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("download: request failed for %s", e.URL)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

func writeResponseAtomic(path string, body io.Reader, maxBytes int64) error {
	if err := fsutil.EnsureParentDir(path, fsutil.DefaultDirPerm); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if dir == "." {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	keepTemp := false
	defer func() {
		_ = tmp.Close()
		if !keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(fsutil.DefaultFilePerm); err != nil {
		return err
	}
	reader := body
	var limited *io.LimitedReader
	if maxBytes > 0 && maxBytes < math.MaxInt64 {
		limited = &io.LimitedReader{R: body, N: maxBytes + 1}
		reader = limited
	}
	if _, err := io.Copy(tmp, reader); err != nil {
		return err
	}
	if limited != nil && limited.N == 0 {
		return ErrResponseTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	keepTemp = true
	return nil
}
