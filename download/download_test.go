package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadManager_ContextCancellation(t *testing.T) {
	// 启动一个模拟的慢速 HTTP 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 模拟慢速下载
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock data"))
	}))
	defer server.Close()

	tmpDir, _ := os.MkdirTemp("", "dm_test")
	defer os.RemoveAll(tmpDir)

	dm := NewDownloadManager(WithWorkers(2))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // 设置极短的超时
	defer cancel()

	_ = dm.Start(ctx)

	savePath := filepath.Join(tmpDir, "test.file")
	err := dm.Add(server.URL, savePath)
	if err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	dm.Wait()
	_ = dm.Close()

	// 验证文件是否未下载成功（或者下载过程被取消）
	if _, err := os.Stat(savePath); err == nil {
		t.Errorf("file should not exist if context was cancelled")
	}
}

func TestDownloadManager_BatchDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	tmpDir, _ := os.MkdirTemp("", "dm_test_batch")
	defer os.RemoveAll(tmpDir)

	dm := NewDownloadManager(WithWorkers(5))
	_ = dm.Start(context.Background())

	const taskCount = 10
	for i := 0; i < taskCount; i++ {
		_ = dm.Add(server.URL, filepath.Join(tmpDir, "file_"+string(rune('0'+i))))
	}

	dm.Wait()
	_ = dm.Close()

	files, _ := os.ReadDir(tmpDir)
	if len(files) != taskCount {
		t.Errorf("expected %d files, got %d", taskCount, len(files))
	}
}

func TestDownloadManager_WithWorkers(t *testing.T) {
	const customWorkers = 12
	dm := NewDownloadManager(WithWorkers(customWorkers))

	if dm.workers != customWorkers {
		t.Errorf("expected workers to be %d, got %d", customWorkers, dm.workers)
	}

	ctx := context.Background()
	_ = dm.Start(ctx)
	defer dm.Close()

	// 间接验证 pool 是否按预期启动
	if dm.pool == nil {
		t.Fatal("pool should not be nil after Start")
	}

	// 注意：dm.pool 的内部 workers 无法直接获取，但我们可以通过 fields 确认配置已正确传递
}
