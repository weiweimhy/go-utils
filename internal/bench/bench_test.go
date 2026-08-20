package bench

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/weiweimhy/go-utils/v6/fsutil"
	"github.com/weiweimhy/go-utils/v6/regexputil"
	"github.com/weiweimhy/go-utils/v6/task"
)

func BenchmarkRegexp(b *testing.B) {
	str := "hello 123 world 456"
	pattern := `(\d+)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = regexputil.MustFindMatches(str, pattern)
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	ctx := context.Background()
	pool := task.NewWorkerPool(
		ctx,
		task.WithWorkerCount(10),
		task.WithBufferSize(100),
		task.WithName("benchmark"),
	)
	defer pool.Close(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.SubmitFunc(func(ctx context.Context) {})
	}
}

func BenchmarkFsutilSave(b *testing.B) {
	data := []byte("benchmark data")
	path := "./bench_tmp/test.file"
	defer os.RemoveAll("./bench_tmp")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fsutil.SaveToFile(path, data)
	}
}
