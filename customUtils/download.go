package customUtils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type DownFileInfo struct {
	SavePath string
	Url      string
	Callback func(url, savePath string, err error) // 下载完成后的回调函数
}

func DownloadFile(url string, path string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	return downloadFile(client, url, path)
}

func downloadFile(client *http.Client, url string, path string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			zap.L().Warn("failed to close http response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	err = CreateDir(path)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			zap.L().Warn("failed to close file", zap.String("path", path), zap.Error(err))
		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	return err
}

// DownloadManager 下载管理器，可以自动启动协程并支持动态添加下载任务
type DownloadManager struct {
	jobs       chan *DownFileInfo
	wg         sync.WaitGroup
	workers    int
	chanSize   int
	delay      time.Duration
	started    bool
	closed     bool
	mu         sync.RWMutex
	once       sync.Once
	onComplete func(url, savePath string, err error)
	ctx        context.Context
	cancel     context.CancelFunc
	eg         *errgroup.Group
}

// DownloadManagerConfig 下载管理器配置
type DownloadManagerConfig struct {
	Workers  int           // Worker 协程数量，默认 20
	ChanSize int           // Channel 缓冲区大小，默认 100
	Delay    time.Duration // 每个下载任务之间的延迟，默认 100ms
	Timeout  time.Duration // HTTP 请求超时时间，默认 60s
}

// NewDownloadManager 创建新的下载管理器
func NewDownloadManager(config *DownloadManagerConfig) *DownloadManager {
	if config == nil {
		config = &DownloadManagerConfig{}
	}

	workers := config.Workers
	if workers <= 0 {
		workers = 20
	}

	chanSize := config.ChanSize
	if chanSize <= 0 {
		chanSize = 100
	}

	delay := config.Delay
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}

	return &DownloadManager{
		jobs:     make(chan *DownFileInfo, chanSize),
		workers:  workers,
		chanSize: chanSize,
		delay:    delay,
		started:  false,
		closed:   false,
	}
}

func (dm *DownloadManager) SetCompletedCallback(callback func(url, savePath string, err error)) {
	dm.onComplete = callback
}

// Start 启动下载管理器，自动启动 worker 协程
func (dm *DownloadManager) Start(ctx context.Context) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.started {
		return nil
	}

	dm.started = true
	dm.ctx, dm.cancel = context.WithCancel(ctx)
	dm.eg, dm.ctx = errgroup.WithContext(dm.ctx)

	for i := 0; i < dm.workers; i++ {
		id := i
		dm.eg.Go(func() error {
			return dm.worker(id)
		})
	}

	zap.L().Info("download manager started", zap.Int("workers", dm.workers))
	return nil
}

// Add 添加单个下载任务
func (dm *DownloadManager) Add(url, savePath string) error {
	return dm.AddWithCallback(url, savePath, nil)
}

// AddWithCallback 添加单个下载任务，并指定回调函数
func (dm *DownloadManager) AddWithCallback(url, savePath string, callback func(url, savePath string, err error)) error {
	dm.mu.RLock()
	closed := dm.closed
	started := dm.started
	dm.mu.RUnlock()

	if closed {
		return fmt.Errorf("download manager is closed")
	}

	if !started {
		if err := dm.Start(context.Background()); err != nil {
			return err
		}
	}

	dm.mu.RLock()
	ctx := dm.ctx
	dm.mu.RUnlock()

	if ctx == nil {
		return fmt.Errorf("download manager context not initialized")
	}

	dm.wg.Add(1)
	select {
	case dm.jobs <- &DownFileInfo{
		SavePath: savePath,
		Url:      url,
		Callback: callback,
	}:
		return nil
	case <-time.After(5 * time.Second):
		dm.wg.Done()
		return fmt.Errorf("add download task timeout, channel may be full")
	case <-ctx.Done():
		dm.wg.Done()
		return ctx.Err()
	}
}

// AddBatch 批量添加下载任务
func (dm *DownloadManager) AddBatch(tasks []*DownFileInfo) error {
	dm.mu.RLock()
	closed := dm.closed
	started := dm.started
	dm.mu.RUnlock()

	if closed {
		return fmt.Errorf("download manager is closed")
	}

	if !started {
		if err := dm.Start(context.Background()); err != nil {
			return err
		}
	}

	dm.mu.RLock()
	ctx := dm.ctx
	dm.mu.RUnlock()

	if ctx == nil {
		return fmt.Errorf("download manager context not initialized")
	}

	for _, task := range tasks {
		dm.wg.Add(1)
		select {
		case dm.jobs <- task:
		case <-time.After(5 * time.Second):
			dm.wg.Done()
			return fmt.Errorf("batch add download tasks timeout")
		case <-ctx.Done():
			dm.wg.Done()
			return ctx.Err()
		}
	}

	return nil
}

// Wait 等待所有下载任务完成
func (dm *DownloadManager) Wait() {
	dm.wg.Wait()
}

// Close 关闭下载管理器，停止接收新任务并等待所有任务完成
func (dm *DownloadManager) Close() error {
	var closeErr error
	dm.once.Do(func() {
		dm.mu.Lock()
		dm.closed = true
		dm.mu.Unlock()

		if dm.cancel != nil {
			dm.cancel()
		}

		close(dm.jobs)
		dm.wg.Wait()

		if dm.eg != nil {
			closeErr = dm.eg.Wait()
		}

		zap.L().Info("download manager closed")
	})
	return closeErr
}

// worker 工作协程
func (dm *DownloadManager) worker(id int) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	dm.mu.RLock()
	ctx := dm.ctx
	dm.mu.RUnlock()

	for {
		select {
		case info, ok := <-dm.jobs:
			if !ok {
				return nil
			}

			zap.L().Info("downloading file",
				zap.Int("worker_id", id),
				zap.String("url", info.Url))

			err := downloadFile(client, info.Url, info.SavePath)
			if err != nil {
				zap.L().Warn("download failed",
					zap.Int("worker_id", id),
					zap.String("url", info.Url),
					zap.Error(err))
			} else {
				zap.L().Info("download completed",
					zap.Int("worker_id", id),
					zap.String("save_path", info.SavePath))
			}

			if info.Callback != nil {
				info.Callback(info.Url, info.SavePath, err)
			}
			if dm.onComplete != nil {
				dm.onComplete(info.Url, info.SavePath, err)
			}

			dm.wg.Done()

			if dm.delay > 0 {
				select {
				case <-time.After(dm.delay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
