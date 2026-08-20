package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weiweimhy/go-utils/v6/fsutil"
	"github.com/weiweimhy/go-utils/v6/httputil"
	"github.com/weiweimhy/go-utils/v6/task"
)

var (
	// ErrManagerClosed is returned when adding work to a closed manager.
	ErrManagerClosed = errors.New("download: manager is closed")
	// ErrManagerNotStarted is returned when adding work before Start.
	ErrManagerNotStarted = errors.New("download: manager not started")
)

// DownloadTask 表示一个下载任务
type DownloadTask struct {
	URL      string
	SavePath string
	Callback func(url, savePath string, err error)
}

// DownloadManager 管理大规模并发下载任务
type DownloadManager struct {
	pool   *task.WorkerPool
	client *http.Client
	delay  time.Duration

	workers    int
	bufferSize int
	log        *slog.Logger

	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    atomic.Bool

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
	return func(dm *DownloadManager) { dm.client = c }
}

// WithLogger sets an optional logger for manager lifecycle and failures.
// When unset, the manager stays silent by default.
func WithLogger(log *slog.Logger) DMOption {
	return func(dm *DownloadManager) { dm.log = log }
}

// dmConfig 内部配置结构
type dmConfig struct {
	workers  int
	chanSize int
	delay    time.Duration
	client   *http.Client
}

func defaultConfig() *dmConfig {
	return &dmConfig{
		workers:  20,
		chanSize: 100,
		client:   httputil.DefaultHTTPClient,
	}
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(opts ...DMOption) *DownloadManager {
	dm := &DownloadManager{
		client:     httputil.DefaultHTTPClient,
		workers:    20,
		bufferSize: 100,
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
	if dm.closed.Load() {
		return ErrManagerClosed
	}
	if dm.pool == nil {
		return ErrManagerNotStarted
	}

	dm.wg.Add(1)

	submitted := dm.pool.SubmitFunc(func(ctx context.Context) {
		defer dm.wg.Done()

		err := downloadFile(ctx, dm.client, url, savePath)
		if err != nil && dm.log != nil {
			dm.log.WarnContext(ctx, "download failed", "url", url, "path", savePath, "error", err)
		}

		if callback != nil {
			callback(url, savePath, err)
		}

		if dm.delay > 0 {
			select {
			case <-time.After(dm.delay):
			case <-ctx.Done():
			}
		}
	})

	if !submitted {
		dm.wg.Done()
		return dm.ctx.Err()
	}

	return nil
}

// Wait 等待所有下载任务完成
func (dm *DownloadManager) Wait() {
	dm.wg.Wait()
}

// Close 关闭下载管理器
func (dm *DownloadManager) Close() error {
	var err error
	dm.closeOnce.Do(func() {
		dm.closed.Store(true)

		// 等待所有任务完成
		dm.wg.Wait()

		if dm.pool != nil {
			if !dm.pool.Close(30 * time.Second) {
				err = fmt.Errorf("download manager close timeout")
			}
		}

		if dm.cancel != nil {
			dm.cancel()
		}
	})
	return err
}

// downloadFile 下载单个文件
func downloadFile(ctx context.Context, client *http.Client, url string, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: status %d", resp.StatusCode)
	}

	if err := writeResponseAtomic(path, resp.Body); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}
	return nil
}

func writeResponseAtomic(path string, body io.Reader) error {
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
	if _, err := io.Copy(tmp, body); err != nil {
		return err
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
