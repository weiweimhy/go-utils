package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weiweimhy/go-utils/v4/errs"
	"github.com/weiweimhy/go-utils/v4/fsutil"
	"github.com/weiweimhy/go-utils/v4/httputil"
	"github.com/weiweimhy/go-utils/v4/task"
	"go.uber.org/zap"
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
	log        *zap.Logger

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
func WithLogger(log *zap.Logger) DMOption {
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
		log:        zap.NewNop(),
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

	dm.log.Info("download manager started",
		zap.Int("workers", workers),
		zap.Int("buffer", bufferSize),
	)
	return nil
}

// Add 添加下载任务
func (dm *DownloadManager) Add(url, savePath string) error {
	return dm.AddWithCallback(url, savePath, nil)
}

// AddWithCallback 添加带回调的下载任务
func (dm *DownloadManager) AddWithCallback(url, savePath string, callback func(url, savePath string, err error)) error {
	if dm.closed.Load() {
		return errs.ErrDownloadManagerClosed
	}
	if dm.pool == nil {
		return errs.ErrDownloadManagerNotStarted
	}

	dm.wg.Add(1)

	submitted := dm.pool.SubmitFunc(func(ctx context.Context) {
		defer dm.wg.Done()

		err := downloadFile(ctx, dm.client, url, savePath)
		if err != nil {
			dm.log.Warn("download failed",
				zap.String("url", url),
				zap.String("path", savePath),
				zap.Error(err),
			)
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

	if err := fsutil.CreateParentDir(path); err != nil {
		return fmt.Errorf("create dir failed: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}
