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
	initialized = false
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

	if !initialized {
		t.Fatal("expected logger to be initialized")
	}
}
