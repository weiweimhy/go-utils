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

func L() *zap.Logger {
	return zap.L()
}

func InitDevelopment() {
	once.Do(func() {
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		)

		logger := zap.New(
			core,
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
			zap.Development(),
		)

		zap.ReplaceGlobals(logger)
	})
}

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
			// 生产环境必须开启 Sampling，防止高并发场景下日志刷屏
			zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
			}),
		)

		zap.ReplaceGlobals(logger)
	})
}

func FromContext(ctx context.Context, fields ...zap.Field) *zap.Logger {
	if ctx == nil {
		return L().With(fields...)
	}
	if l, ok := ctx.Value(ctxKey).(*zap.Logger); ok && l != nil {
		return l.With(fields...)
	}
	return L().With(fields...)
}

func ToContext(ctx context.Context, l *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey, l)
}

type CtxLogger struct {
	Ctx context.Context
	Log *zap.Logger
}

// NewCtxLogger 基于给定 Ctx 获取/创建 logger，并将其写回新的 Ctx
func NewCtxLogger(ctx context.Context, fields ...zap.Field) CtxLogger {
	log := FromContext(ctx, fields...)
	newCtx := ToContext(ctx, log)
	return CtxLogger{
		Ctx: newCtx,
		Log: log,
	}
}

// With 为当前 CtxLogger 附加字段，并返回新的 CtxLogger（同时更新 Ctx 中的 logger）
func (cl CtxLogger) With(fields ...zap.Field) CtxLogger {
	log := cl.Log.With(fields...)
	ctx := ToContext(cl.Ctx, log)
	return CtxLogger{
		Ctx: ctx,
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
