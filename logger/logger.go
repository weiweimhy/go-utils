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

// Options 包含 Logger 的所有配置项
type Options struct {
	Filename    string
	Level       zapcore.Level
	MaxMB       int
	MaxBackups  int
	MaxAge      int
	Compress    bool
	Sampling    *zap.SamplingConfig
	Development bool
}

type Option func(*Options)

// DefaultOptions 默认设置
func DefaultOptions() *Options {
	return &Options{
		Filename:    "./logs/app.log",
		Level:       zap.InfoLevel,
		MaxMB:       100,
		MaxBackups:  10,
		MaxAge:      30,
		Compress:    true,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
	}
}

func WithFilename(f string) Option     { return func(o *Options) { o.Filename = f } }
func WithLevel(l zapcore.Level) Option { return func(o *Options) { o.Level = l } }
func WithDevelopment(d bool) Option    { return func(o *Options) { o.Development = d } }

// Init 使用 Functional Options 初始化全局 Logger
func Init(opts ...Option) {
	once.Do(func() {
		o := DefaultOptions()
		for _, opt := range opts {
			opt(o)
		}

		writeSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   o.Filename,
			MaxSize:    o.MaxMB,
			MaxBackups: o.MaxBackups,
			MaxAge:     o.MaxAge,
			Compress:   o.Compress,
			LocalTime:  true,
		})

		var encoderConfig zapcore.EncoderConfig
		if o.Development {
			encoderConfig = zap.NewDevelopmentEncoderConfig()
		} else {
			encoderConfig = zap.NewProductionEncoderConfig()
			encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		}

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), writeSyncer),
			zap.NewAtomicLevelAt(o.Level),
		)

		zapOpts := []zap.Option{
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		}

		if o.Sampling != nil {
			zapOpts = append(zapOpts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewSamplerWithOptions(c, time.Second, o.Sampling.Initial, o.Sampling.Thereafter)
			}))
		}

		logger := zap.New(core, zapOpts...)
		zap.ReplaceGlobals(logger)
	})
}

// L 返回全局共享的 Logger 实例
func L() *zap.Logger {
	return zap.L()
}

// FromContext 从 context 获取 Logger，由于重构了 Global 模式，此处建议作为首选
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

// Trace 自动记录函数耗时与 Panic 捕获
func Trace(log *zap.Logger, funcName string, fields ...zap.Field) func() {
	start := time.Now()
	log.Debug("→ function entry", append(fields, zap.String("func", funcName))...)

	return func() {
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

// InvalidParam 统一参数错误处理
func InvalidParam(log *zap.Logger, msg string, fields ...zap.Field) error {
	log.Error("invalid parameter",
		append(fields,
			zap.String("error", "invalid_param"),
			zap.String("func", getCallerFuncName()),
			zap.Stack("stack"),
		)...,
	)
	return fmt.Errorf("invalid param: %s", msg)
}

var funcNameCache sync.Map

func getCallerFuncName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	if name, ok := funcNameCache.Load(pc); ok {
		return name.(string)
	}

	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}

	fullName := f.Name()
	parts := strings.Split(fullName, ".")
	short := parts[len(parts)-1]
	if strings.HasPrefix(short, "(") && len(parts) >= 2 {
		short = parts[len(parts)-2] + "." + short
	}

	_, line := f.FileLine(pc)
	result := fmt.Sprintf("%s:%d", short, line)
	funcNameCache.Store(pc, result)
	return result
}
