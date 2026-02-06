// Package logger provides structured logging with context propagation and observability support.
package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/config"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
)

type ctxKey struct{}

// Standard attribute keys for observability.
const (
	KeyService       = "service"
	KeyFunction      = "fn"
	KeyError         = "error"
	KeyRequestID     = "request_id"
	KeyCorrelationID = "correlation_id"
	KeyBatchID       = "batch_id"
	KeyEventID       = "event_id"
	KeyEventType     = "event_type"
	KeyAccountID     = "account_id"
	KeyStatus        = "status"
	KeyDuration      = "duration_ms"
	KeyAttempts      = "attempts"
	KeyErrorCode     = "error_code"
	KeyRetryable     = "retryable"
	KeySource        = "source"
	KeyCount         = "count"
)

// Config holds logger configuration.
type Config struct {
	Output      io.Writer
	ServiceName string
	Level       slog.Level
	AddSource   bool
}

// DefaultConfig returns sensible defaults for production.
func DefaultConfig() Config {
	level := slog.LevelInfo
	if config.GetEnvBool(constants.EnvLogLevelDebug, false) {
		level = slog.LevelDebug
	}

	return Config{
		ServiceName: config.GetEnv(constants.EnvServiceName, ""),
		Level:       level,
		Output:      os.Stdout,
		AddSource:   true,
	}
}

// Init initializes the global logger with the given configuration.
func Init(cfg Config) {
	h := slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	})

	base := slog.New(h)
	if cfg.ServiceName != "" {
		base = base.With(slog.String(KeyService, cfg.ServiceName))
	}

	slog.SetDefault(base)
	log.SetFlags(0)
	log.SetOutput(&stdLogWriter{base})
}

// InitDefault initializes with default configuration. Call once in main.
func InitDefault() {
	Init(DefaultConfig())
}

type stdLogWriter struct {
	logger *slog.Logger
}

func (w *stdLogWriter) Write(p []byte) (int, error) {
	w.logger.Info(strings.TrimSuffix(string(p), "\n"))

	return len(p), nil
}

// With adds attributes to the logger in context.
func With(ctx context.Context, args ...any) context.Context {
	return context.WithValue(ctx, ctxKey{}, From(ctx).With(args...))
}

// From retrieves the logger from context, or returns the default logger.
func From(ctx context.Context) *slog.Logger {
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
	From(ctx).InfoContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// Error logs at ERROR level with automatic caller detection.
func Error(ctx context.Context, msg string, err error, args ...any) {
	From(ctx).ErrorContext(
		ctx,
		msg,
		append(args, slog.String(KeyFunction, caller()), slog.String(KeyError, err.Error()))...,
	)
}

// Debug logs at DEBUG level with automatic caller detection.
func Debug(ctx context.Context, msg string, args ...any) {
	From(ctx).DebugContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// Warn logs at WARN level with automatic caller detection.
func Warn(ctx context.Context, msg string, args ...any) {
	From(ctx).WarnContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// EventReceived logs when an event is received.
func EventReceived(ctx context.Context, eventID, eventType, accountID, batchID string) {
	Info(ctx, "event_received",
		slog.String(KeyEventID, eventID),
		slog.String(KeyEventType, eventType),
		slog.String(KeyAccountID, accountID),
		slog.String(KeyBatchID, batchID),
	)
}

// EventProcessed logs when an event is successfully processed.
func EventProcessed(ctx context.Context, eventID string, durationMs int64) {
	Info(ctx, "event_processed",
		slog.String(KeyEventID, eventID),
		slog.String(KeyStatus, "PROCESSED"),
		slog.Int64(KeyDuration, durationMs),
	)
}

// EventFailed logs when an event processing fails.
func EventFailed(ctx context.Context, eventID, errorCode string, attempts int, retryable bool) {
	Warn(ctx, "event_failed",
		slog.String(KeyEventID, eventID),
		slog.String(KeyStatus, "FAILED"),
		slog.String(KeyErrorCode, errorCode),
		slog.Int(KeyAttempts, attempts),
		slog.Bool(KeyRetryable, retryable),
	)
}

// EventRetried logs when an event is scheduled for retry.
func EventRetried(ctx context.Context, eventID, errorCode string, attempts int) {
	Warn(ctx, "event_retried",
		slog.String(KeyEventID, eventID),
		slog.String(KeyStatus, "RETRIED"),
		slog.String(KeyErrorCode, errorCode),
		slog.Int(KeyAttempts, attempts),
	)
}

// BatchReceived logs when a batch of events is received.
func BatchReceived(ctx context.Context, batchID, source string, count int) {
	Info(ctx, "batch_received",
		slog.String(KeyBatchID, batchID),
		slog.String(KeySource, source),
		slog.Int(KeyCount, count),
	)
}

// BatchCompleted logs batch processing completion.
func BatchCompleted(ctx context.Context, batchID string, processed, failed int, durationMs int64) {
	Info(ctx, "batch_completed",
		slog.String(KeyBatchID, batchID),
		slog.Int("processed_count", processed),
		slog.Int("failed_count", failed),
		slog.Int64(KeyDuration, durationMs),
	)
}

// OperationStart returns a function to log operation duration.
//
// Usage:
//
//	done := logger.OperationStart(ctx, "save_events")
//	defer done()
func OperationStart(ctx context.Context, operation string) func() {
	start := time.Now()

	Debug(ctx, "operation_start", slog.String("operation", operation))

	return func() {
		Debug(ctx, "operation_end",
			slog.String("operation", operation),
			slog.Int64(KeyDuration, time.Since(start).Milliseconds()),
		)
	}
}

// WithRequestContext enriches context with common request attributes.
func WithRequestContext(
	ctx context.Context,
	requestID, correlationID, path string,
) context.Context {
	return With(ctx,
		slog.String(KeyRequestID, requestID),
		slog.String(KeyCorrelationID, correlationID),
		slog.String("path", path),
	)
}

// WithEventContext enriches context with event-specific attributes.
func WithEventContext(ctx context.Context, eventID, batchID string) context.Context {
	return With(ctx,
		slog.String(KeyEventID, eventID),
		slog.String(KeyBatchID, batchID),
	)
}
