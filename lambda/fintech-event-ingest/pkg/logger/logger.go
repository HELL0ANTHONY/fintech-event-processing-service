// Package logger provides structured logging with context propagation.
package logger

import (
	"context"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

type ctxKey struct{}

// Init initializes the global logger. Call once in main.
func Init() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		level = slog.LevelDebug
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	base := slog.New(h)

	if svc := os.Getenv("SERVICE_NAME"); svc != "" {
		base = base.With(slog.String("service", svc))
	}

	slog.SetDefault(base)
	log.SetFlags(0)
	log.SetOutput(&writer{})
}

type writer struct{}

func (w *writer) Write(p []byte) (int, error) {
	slog.Info(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

// With adds attributes to the logger in context.
func With(ctx context.Context, args ...any) context.Context {
	return context.WithValue(ctx, ctxKey{}, from(ctx).With(args...))
}

func from(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}

	return slog.Default()
}

func caller() string {
	pc, _, _, ok := runtime.Caller(2)

	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc).Name()

	if idx := strings.LastIndex(fn, "/"); idx != -1 {
		fn = fn[idx+1:]
	}

	return fn
}

// Info logs at INFO level with automatic caller detection.
func Info(ctx context.Context, msg string, args ...any) {
	from(ctx).Info(msg, append(args, slog.String("fn", caller()))...)
}

// Error logs at ERROR level with automatic caller detection.
func Error(ctx context.Context, msg string, err error, args ...any) {
	from(
		ctx,
	).Error(msg, append(args, slog.String("fn", caller()), slog.String("error", err.Error()))...)
}

// Debug logs at DEBUG level with automatic caller detection.
func Debug(ctx context.Context, msg string, args ...any) {
	from(ctx).Debug(msg, append(args, slog.String("fn", caller()))...)
}

// Warn logs at WARN level with automatic caller detection.
func Warn(ctx context.Context, msg string, args ...any) {
	from(ctx).Warn(msg, append(args, slog.String("fn", caller()))...)
}

// ErrorAttrs logs at ERROR level with automatic caller detection and additional attributes.
func ErrorAttrs(ctx context.Context, msg string, args ...any) {
	from(ctx).Error(msg, append(args, slog.String("fn", caller()))...)
}
