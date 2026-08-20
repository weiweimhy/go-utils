package logger

import (
	"context"
	"os"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func resetLoggerForTest() {
	once = sync.Once{}
	zap.ReplaceGlobals(zap.NewNop())
}

func TestToContextFromContext(t *testing.T) {
	resetLoggerForTest()

	base := zap.NewNop().With(zap.String("component", "test"))
	ctx := ToContext(context.Background(), base)
	got := FromContext(ctx)
	if got == nil {
		t.Fatal("expected logger from context")
	}
}

func TestInitWritesLogFile(t *testing.T) {
	resetLoggerForTest()

	Init(WithFilename(os.DevNull))
	L().Info("hello")
	_ = L().Sync()

}

func TestConcurrentInitIsIdempotent(t *testing.T) {
	resetLoggerForTest()
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Init(WithFilename(os.DevNull))
		}()
	}
	wg.Wait()
	if L() == nil {
		t.Fatal("expected initialized logger")
	}
}

func TestCtxLoggerWithHandlesZeroValue(t *testing.T) {
	resetLoggerForTest()
	updated := (CtxLogger{}).With(zap.String("key", "value"))
	if updated.Log == nil {
		t.Fatal("With() should use the global logger for a zero-value CtxLogger")
	}
	if updated.Ctx == nil {
		t.Fatal("With() should create a context for a zero-value CtxLogger")
	}
}

func TestLoggerHelpersHandleNilLogger(t *testing.T) {
	resetLoggerForTest()
	defer Trace(nil, "test")()
	if err := InvalidParam(nil, "bad input"); err == nil {
		t.Fatal("InvalidParam() should return an error")
	}
}
