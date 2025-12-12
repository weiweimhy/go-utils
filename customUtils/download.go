package customUtils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/weiweimhy/go-utils/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type DownFileInfo struct {
	SavePath string
	Url      string
	Callback func(url, savePath string, err error)
}

// DownloadManager 管理大规模并发下载任务
type DownloadManager struct {
	jobs chan *DownFileInfo
	wg   sync.WaitGroup
	mu   sync.RWMutex
	once sync.Once

	workers  int
	chanSize int
	delay    time.Duration
	client   *http.Client

	started    bool
	closed     bool
	onComplete func(url, savePath string, err error)
	ctx        context.Context
	cancel     context.CancelFunc
	eg         *errgroup.Group
}

type DMOption func(*DownloadManager)

func WithWorkers(w int) DMOption         { return func(dm *DownloadManager) { dm.workers = w } }
func WithChanSize(s int) DMOption        { return func(dm *DownloadManager) { dm.chanSize = s } }
func WithDelay(d time.Duration) DMOption { return func(dm *DownloadManager) { dm.delay = d } }
func WithClient(c *http.Client) DMOption { return func(dm *DownloadManager) { dm.client = c } }

// NewDownloadManager 使用 Functional Options 创建管理器
func NewDownloadManager(opts ...DMOption) *DownloadManager {
	dm := &DownloadManager{
		workers:  20,
		chanSize: 100,
		client:   DefaultHttpClient,
	}
	for _, opt := range opts {
		opt(dm)
	}
	dm.jobs = make(chan *DownFileInfo, dm.chanSize)
	return dm
}

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

	logger.L().Info("download manager started", zap.Int("workers", dm.workers))
	return nil
}

func (dm *DownloadManager) Add(url, savePath string) error {
	dm.mu.RLock()
	if dm.closed {
		dm.mu.RUnlock()
		return fmt.Errorf("manager is closed")
	}
	if !dm.started {
		dm.mu.RUnlock()
		return fmt.Errorf("manager not started")
	}
	ctx := dm.ctx
	dm.mu.RUnlock()

	dm.wg.Add(1)
	select {
	case dm.jobs <- &DownFileInfo{SavePath: savePath, Url: url}:
		return nil
	case <-ctx.Done():
		dm.wg.Done()
		return ctx.Err()
	}
}

func (dm *DownloadManager) Wait() {
	dm.wg.Wait()
}

func (dm *DownloadManager) Close() error {
	dm.once.Do(func() {
		dm.mu.Lock()
		dm.closed = true
		dm.mu.Unlock()

		if dm.cancel != nil {
			dm.cancel()
		}
		close(dm.jobs)
	})

	dm.wg.Wait()
	if dm.eg != nil {
		return dm.eg.Wait()
	}
	return nil
}

func (dm *DownloadManager) worker(id int) error {
	for {
		select {
		case info, ok := <-dm.jobs:
			if !ok {
				return nil
			}

			err := downloadFile(dm.ctx, dm.client, info.Url, info.SavePath)
			if err != nil {
				logger.L().Warn("download failed", zap.String("url", info.Url), zap.Error(err))
			}

			if dm.onComplete != nil {
				dm.onComplete(info.Url, info.SavePath, err)
			}
			dm.wg.Done()

			if dm.delay > 0 {
				select {
				case <-time.After(dm.delay):
				case <-dm.ctx.Done():
					return dm.ctx.Err()
				}
			}
		case <-dm.ctx.Done():
			return dm.ctx.Err()
		}
	}
}

func downloadFile(ctx context.Context, client *http.Client, url string, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: %d", resp.StatusCode)
	}

	if err := CreateDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
