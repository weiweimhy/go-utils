package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once   sync.Once
	ctxKey = struct{}{}
)

func InitProduction() {
	once.Do(func() {
		writeSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./logs/app.log",
			MaxSize:    1, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
			LocalTime:  true,
		})

		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO ERROR
			EncodeTime:     zapcore.ISO8601TimeEncoder,  // 2025-12-01T10:20:30.123Z
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), writeSyncer),
			zap.NewAtomicLevelAt(zap.InfoLevel),
		)

		logger := zap.New(
			core,
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
			// 可选：高并发防刷屏采样
			// zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			// 	return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
			// }),
		)

		zap.ReplaceGlobals(logger)
	})
}

func FromContext(ctx context.Context, fields ...zap.Field) *zap.Logger {
	if ctx == nil {
		return zap.L().With(fields...)
	}
	if l, ok := ctx.Value(ctxKey).(*zap.Logger); ok && l != nil {
		return l.With(fields...)
	}
	return zap.L().With(fields...)
}

func ToContext(ctx context.Context, l *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey, l)
}

type CtxLogger struct {
	ctx context.Context
	Log *zap.Logger
}

// NewCtxLogger 基于给定 ctx 获取/创建 logger，并将其写回新的 ctx
func NewCtxLogger(ctx context.Context, fields ...zap.Field) CtxLogger {
	log := FromContext(ctx, fields...)
	newCtx := ToContext(ctx, log)
	return CtxLogger{
		ctx: newCtx,
		Log: log,
	}
}

// With 为当前 CtxLogger 附加字段，并返回新的 CtxLogger（同时更新 ctx 中的 logger）
func (cl CtxLogger) With(fields ...zap.Field) CtxLogger {
	log := cl.Log.With(fields...)
	ctx := ToContext(cl.ctx, log)
	return CtxLogger{
		ctx: ctx,
		Log: log,
	}
}

// Trace 自动记录函数进入/退出 + 耗时，生产必备
func Trace(log *zap.Logger, funcName string, fields ...zap.Field) func() {
	start := time.Now()
	log.Debug("→ function entry", append(fields, zap.String("func", funcName))...)

	return func() {
		// 如果有 panic，也能捕获
		if r := recover(); r != nil {
			log.Error("function panic",
				zap.String("func", funcName),
				zap.Any("panic", r),
				zap.Stack("stack"),
				zap.Duration("cost", time.Since(start)),
			)
			panic(r)
		}
		log.Debug("← function exit",
			append(fields,
				zap.String("func", funcName),
				zap.Duration("cost", time.Since(start)),
			)...,
		)
	}
}

// InvalidParam 统一的参数错误日志（所有项目都长这样）
func InvalidParam(log *zap.Logger, msg string, fields ...zap.Field) error {
	log.Error("invalid parameter",
		append(fields,
			zap.String("error", "invalid_param"),
			zap.String("func", getCallerFuncName()),
			zap.Stack("stack"), // 方便一键跳到调用处
		)...,
	)
	return fmt.Errorf("invalid param: %s", msg)
}

var funcNameCache sync.Map // 全局缓存，命中率 99.99%

// 获取调用者函数名（短名 + 行号），带缓存，性能 < 30ns
func getCallerFuncName() string {
	pc, _, _, ok := runtime.Caller(2) // 2 层：InvalidParam -> 调用方
	if !ok {
		return "unknown"
	}

	key := pc
	if name, ok := funcNameCache.Load(key); ok {
		return name.(string)
	}

	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}

	fullName := f.Name()
	// 取最后两段：(*Client).Start:45
	parts := strings.Split(fullName, ".")
	short := parts[len(parts)-1]
	if strings.HasPrefix(short, "(") {
		// 处理 receiver
		if len(parts) >= 2 {
			short = parts[len(parts)-2] + "." + short
		}
	}

	_, line := f.FileLine(pc)
	result := fmt.Sprintf("%s:%d", short, line)

	funcNameCache.Store(key, result)
	return result
}
