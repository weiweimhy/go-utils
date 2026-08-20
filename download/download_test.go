package download

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadManager_ContextCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
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
	<-requestStarted

	dm.Wait()
	_ = dm.Close()

	// 验证文件是否未下载成功（或者下载过程被取消）
	if _, err := os.Stat(savePath); err == nil {
		t.Errorf("file should not exist if context was cancelled")
	}
}

func TestDownloadManagerCloseCancelsRunningDownload(t *testing.T) {
	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
	}))
	defer server.Close()

	dm := NewDownloadManager()
	if err := dm.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	result := make(chan error, 1)
	if err := dm.AddWithCallback(server.URL, filepath.Join(t.TempDir(), "download"), func(url, savePath string, err error) {
		result <- err
	}); err != nil {
		t.Fatalf("AddWithCallback() error = %v", err)
	}
	<-requestStarted

	if err := dm.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("download error = %v, want context cancellation", err)
	}
}

func TestDownloadManagerCloseFromCallbackDoesNotDeadlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	dm := NewDownloadManager(WithWorkers(1))
	if err := dm.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	closedFromCallback := make(chan error, 1)
	if err := dm.AddWithCallback(server.URL, filepath.Join(t.TempDir(), "download"), func(url, savePath string, err error) {
		closedFromCallback <- dm.Close()
	}); err != nil {
		t.Fatalf("AddWithCallback() error = %v", err)
	}

	select {
	case err := <-closedFromCallback:
		if err != nil {
			t.Fatalf("Close() from callback error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() from callback deadlocked")
	}
	if err := dm.Close(); err != nil {
		t.Fatalf("external Close() error = %v", err)
	}
}

func TestDownloadManagerLimitsResponseSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("oversized"))
	}))
	defer server.Close()

	dm := NewDownloadManager(WithMaxBytes(3))
	if err := dm.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer dm.Close()
	result := make(chan error, 1)
	path := filepath.Join(t.TempDir(), "download")
	if err := dm.AddWithCallback(server.URL, path, func(url, savePath string, err error) {
		result <- err
	}); err != nil {
		t.Fatalf("AddWithCallback() error = %v", err)
	}
	dm.Wait()
	if err := <-result; !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("download error = %v, want ErrResponseTooLarge", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("oversized download left output file, stat error = %v", err)
	}
}

func TestWriteResponseAtomicAcceptsMaxInt64Limit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "download")
	if err := writeResponseAtomic(path, strings.NewReader("ok"), math.MaxInt64); err != nil {
		t.Fatalf("writeResponseAtomic() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(data); got != "ok" {
		t.Fatalf("file content = %q, want %q", got, "ok")
	}
}

func TestDownloadManagerRejectsStartAfterClose(t *testing.T) {
	dm := NewDownloadManager()
	if err := dm.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := dm.Start(context.Background()); !errors.Is(err, ErrManagerClosed) {
		t.Fatalf("Start() error = %v, want ErrManagerClosed", err)
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

func TestDownloadManager_StartWithNilContext(t *testing.T) {
	dm := NewDownloadManager()

	if err := dm.Start(nil); err != nil {
		t.Fatalf("expected nil context to be accepted, got error: %v", err)
	}
	defer dm.Close()

	if dm.pool == nil {
		t.Fatal("pool should not be nil after Start(nil)")
	}
	if dm.ctx == nil {
		t.Fatal("manager context should be initialized")
	}
}
