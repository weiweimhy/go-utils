package httputil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/weiweimhy/go-utils/v2/task"
)

// DownloadTask 实现 Task 接口，用于 WorkerPool
type DownloadTask struct {
	URL      string
	SavePath string
	Callback func(url, savePath string, err error)
	Client   *http.Client
}

func (dt *DownloadTask) Execute() {
	client := dt.Client
	if client == nil {
		client = defaultDownloadClient
	}

	err := DownloadFileWithClient(client, dt.URL, dt.SavePath)

	if dt.Callback != nil {
		dt.Callback(dt.URL, dt.SavePath, err)
	}
}

// NewDownloadTask 创建下载任务，使用默认共享 Client
func NewDownloadTask(url, savePath string, callback func(url, savePath string, err error)) *DownloadTask {
	return &DownloadTask{
		URL:      url,
		SavePath: savePath,
		Callback: callback,
	}
}

// NewDownloadTaskWithClient 创建下载任务，使用自定义 Client
func NewDownloadTaskWithClient(client *http.Client, url, savePath string, callback func(url, savePath string, err error)) *DownloadTask {
	return &DownloadTask{
		URL:      url,
		SavePath: savePath,
		Callback: callback,
		Client:   client,
	}
}

// DownloadBatch 批量下载文件，使用默认共享 Client
func DownloadBatch(pool *task.WorkerPool, tasks []*DownloadTask) error {
	return DownloadBatchWithClient(pool, nil, tasks)
}

// DownloadBatchWithClient 批量下载文件，使用指定 Client
func DownloadBatchWithClient(pool *task.WorkerPool, client *http.Client, tasks []*DownloadTask) error {
	if pool == nil {
		return fmt.Errorf("worker pool is nil")
	}

	for _, dt := range tasks {
		if client != nil && dt.Client == nil {
			dt.Client = client
		}
		if !pool.Submit(dt) {
			return fmt.Errorf("failed to submit download task: %s", dt.URL)
		}
	}

	return nil
}

// NewDownloadClient 创建自定义超时的下载 Client
func NewDownloadClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = defaultDownloadTimeout
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
