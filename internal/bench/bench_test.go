package bench

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/weiweimhy/go-utils/v3/fsutil"
	"github.com/weiweimhy/go-utils/v3/jwt"
	"github.com/weiweimhy/go-utils/v3/logger"
	"github.com/weiweimhy/go-utils/v3/regexputil"
	"github.com/weiweimhy/go-utils/v3/task"
	"go.uber.org/zap"
)

func BenchmarkLogger(b *testing.B) {
	logger.Init(logger.WithFilename("./bench.log"))
	l := logger.L()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message", zap.Int("index", i))
	}
}

func BenchmarkRegexp(b *testing.B) {
	str := "hello 123 world 456"
	pattern := `(\d+)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = regexputil.GetRegexpMatches(str, pattern)
	}
}

func BenchmarkJWT(b *testing.B) {
	j, _ := jwt.NewJWT(jwt.WithSecret("bench-secret-key-256-bits-long!!"))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokens, _ := j.Generate(ctx, "user123", nil)
		_, _ = j.Validate(ctx, tokens.AccessToken)
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
