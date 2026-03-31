package perf

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	errSample = errors.New("something went wrong")

	// pre-built loggers pointing at io.Discard to eliminate I/O noise
	slogLogger    = slog.New(slog.NewJSONHandler(io.Discard, nil))
	slogWarnOnly  = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	zlogLogger    = zerolog.New(io.Discard)
	zapLogger     = newZapLogger(zapcore.InfoLevel)
	zapWarnLogger = newZapLogger(zapcore.WarnLevel)
	zapSugar      = zapLogger.Sugar()
)

func newZapLogger(level zapcore.Level) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(io.Discard),
		level,
	)
	return zap.New(core)
}

// ---------------------------------------------------------------------------
// 场景 1：纯文字消息，无字段
// ---------------------------------------------------------------------------

func BenchmarkSimple(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			slogLogger.Info("benchmark message")
		}
	})

	b.Run("zerolog", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			zlogLogger.Info().Msg("benchmark message")
		}
	})

	b.Run("zap", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			zapLogger.Info("benchmark message")
		}
	})
}

// ---------------------------------------------------------------------------
// 场景 2：结构化字段（string + int + error）
// ---------------------------------------------------------------------------

func BenchmarkWithFields(b *testing.B) {
	b.Run("slog/Info", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			slogLogger.Info("benchmark message",
				"url", "https://example.com/api",
				"status", 200,
				"err", errSample,
			)
		}
	})

	// LogAttrs 是 slog 内部最优路径，避免 key-value any 装箱
	b.Run("slog/LogAttrs", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			slogLogger.LogAttrs(context.TODO(), slog.LevelInfo, "benchmark message",
				slog.String("url", "https://example.com/api"),
				slog.Int("status", 200),
				slog.Any("err", errSample),
			)
		}
	})

	b.Run("zerolog", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			zlogLogger.Info().
				Str("url", "https://example.com/api").
				Int("status", 200).
				Err(errSample).
				Msg("benchmark message")
		}
	})

	b.Run("zap/typed", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			zapLogger.Info("benchmark message",
				zap.String("url", "https://example.com/api"),
				zap.Int("status", 200),
				zap.Error(errSample),
			)
		}
	})

	// Sugar 使用 interface{} 变参，与 slog 类似会产生装箱开销
	b.Run("zap/sugared", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			zapSugar.Infow("benchmark message",
				"url", "https://example.com/api",
				"status", 200,
				"err", errSample,
			)
		}
	})
}

// ---------------------------------------------------------------------------
// 场景 3：级别被过滤（logger 阈值高于调用级别，实际不输出）
// 体现"零成本抽象"：zerolog/zap 承诺此时零分配
// ---------------------------------------------------------------------------

func BenchmarkDisabledLevel(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			// slogWarnOnly 的级别为 Warn，Info 被过滤
			slogWarnOnly.Info("benchmark message",
				"url", "https://example.com/api",
				"status", 200,
				"err", errSample,
			)
		}
	})

	b.Run("zerolog", func(b *testing.B) {
		// 将 zerolog 全局级别暂时调高，结束后恢复
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			zlogLogger.Info().
				Str("url", "https://example.com/api").
				Int("status", 200).
				Err(errSample).
				Msg("benchmark message")
		}
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	})

	b.Run("zap", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			// zapWarnLogger 的级别为 Warn，Info 被过滤
			zapWarnLogger.Info("benchmark message",
				zap.String("url", "https://example.com/api"),
				zap.Int("status", 200),
				zap.Error(errSample),
			)
		}
	})
}
