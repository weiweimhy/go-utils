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
	Timeout  time.Duration
}

func (dt *DownloadTask) Execute() {
	if dt.Client == nil {
		timeout := dt.Timeout
		if timeout <= 0 {
			timeout = defaultDownloadTimeout
		}
		dt.Client = &http.Client{
			Timeout: timeout,
		}
	}

	err := downloadFile(dt.Client, dt.URL, dt.SavePath)

	if dt.Callback != nil {
		dt.Callback(dt.URL, dt.SavePath, err)
	}
}

// NewDownloadTask 创建下载任务，使用默认超时时间
func NewDownloadTask(url, savePath string, callback func(url, savePath string, err error)) *DownloadTask {
	return NewDownloadTaskWithTimeout(url, savePath, defaultDownloadTimeout, callback)
}

// NewDownloadTaskWithTimeout 创建下载任务，可自定义超时时间
func NewDownloadTaskWithTimeout(url, savePath string, timeout time.Duration, callback func(url, savePath string, err error)) *DownloadTask {
	return &DownloadTask{
		URL:      url,
		SavePath: savePath,
		Callback: callback,
		Timeout:  timeout,
	}
}

// DownloadBatch 批量下载文件（使用 WorkerPool + Task 模式）
// 这是推荐的批量下载方式，符合项目规范
func DownloadBatch(pool *task.WorkerPool, tasks []*DownloadTask) error {
	if pool == nil {
		return fmt.Errorf("worker pool is nil")
	}

	for _, dt := range tasks {
		if !pool.Submit(dt) {
			return fmt.Errorf("failed to submit download task: %s", dt.URL)
		}
	}

	return nil
}
