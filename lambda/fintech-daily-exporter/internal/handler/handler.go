package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-daily-exporter/internal/exporter"
)

// EventBridgeEvent represents the incoming EventBridge scheduled event.
type EventBridgeEvent struct {
	Version    string                 `json:"version"`
	ID         string                 `json:"id"`
	DetailType string                 `json:"detail_type"`
	Source     string                 `json:"source"`
	Account    string                 `json:"account"`
	Time       string                 `json:"time"`
	Region     string                 `json:"region"`
	Detail     EventBridgeEventDetail `json:"detail"`
}

// EventBridgeEventDetail contains the export configuration.
type EventBridgeEventDetail struct {
	ExportDate   string   `json:"export_date"`
	Format       string   `json:"format"`
	StatusFilter []string `json:"status_filter"`
}

// Handler processes EventBridge scheduled events.
type Handler struct {
	exporter exporter.Exporter
}

// New creates a new Handler.
func New(e exporter.Exporter) *Handler {
	return &Handler{exporter: e}
}

// Handle is the Lambda entry point for EventBridge events.
func (h *Handler) Handle(
	ctx context.Context,
	event json.RawMessage,
) (*models.ExportResponse, error) {
	start := time.Now()

	var ebEvent EventBridgeEvent
	if err := json.Unmarshal(event, &ebEvent); err != nil {
		logger.Error(ctx, "failed to parse EventBridge event", err)
		return nil, err
	}

	ctx = logger.With(ctx,
		"event_id", ebEvent.ID,
		"export_date", ebEvent.Detail.ExportDate,
	)
	logger.Info(ctx, "starting daily export")

	exportDate := ebEvent.Detail.ExportDate
	if exportDate == "" {
		exportDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	statusFilter := ebEvent.Detail.StatusFilter
	if len(statusFilter) == 0 {
		statusFilter = []string{"PROCESSED", "FAILED"}
	}

	response, err := h.exporter.Export(ctx, exportDate, statusFilter)
	if err != nil {
		logger.Error(ctx, "export failed", err)
		return nil, err
	}

	response.DurationMs = time.Since(start).Milliseconds()
	logger.Info(ctx, "export completed",
		"exported_count", response.ExportedCount,
		"failed_count", response.FailedCount,
		"duration_ms", response.DurationMs,
	)

	return response, nil
}
