// Package exporter handles the export of events to S3.
package exporter

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"

	s3repo "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-daily-exporter/internal/repository"
)

// Exporter handles the export of events to S3.
type Exporter interface {
	Export(ctx context.Context, date string, statusFilter []string) (*models.ExportResponse, error)
}

type exporter struct {
	eventRepo repository.EventRepository
	s3Repo    s3repo.S3Repository
}

// New creates a new Exporter.
func New(eventRepo repository.EventRepository, s3Repo s3repo.S3Repository) Exporter {
	return &exporter{
		eventRepo: eventRepo,
		s3Repo:    s3Repo,
	}
}

// Export queries events by status and date, generates a CSV, and uploads to S3.
func (e *exporter) Export(
	ctx context.Context,
	date string,
	statusFilter []string,
) (*models.ExportResponse, error) {
	response := &models.ExportResponse{
		ExportDate:   date,
		StatusFilter: statusFilter,
	}

	var allEvents []*models.DynamoDBEvent

	for _, status := range statusFilter {
		// Pass 0 for no limit.
		events, err := e.eventRepo.QueryEventsByStatus(ctx, status, date, 0)
		if err != nil {
			logger.Error(ctx, "failed to query events", err, "status", status)

			return response, err
		}

		allEvents = append(allEvents, events...)
	}

	logger.Info(ctx, "queried events for export", "total_count", len(allEvents))

	if len(allEvents) == 0 {
		logger.Info(ctx, "no events to export")

		return response, nil
	}

	csvData, err := e.generateCSV(allEvents)
	if err != nil {
		return response, fmt.Errorf("generate CSV: %w", err)
	}

	// Format: exports/2025-01/events-2025-01-15.csv
	key := fmt.Sprintf("exports/%s/events-%s.csv", date[:7], date)

	s3Info, err := e.s3Repo.Upload(ctx, key, csvData)
	if err != nil {
		return response, fmt.Errorf("upload to S3: %w", err)
	}

	response.ExportedCount = countByStatus(allEvents, constants.StatusProcessed)
	response.FailedCount = countByStatus(allEvents, constants.StatusFailed)
	response.S3 = s3Info

	return response, nil
}

func (e *exporter) generateCSV(events []*models.DynamoDBEvent) ([]byte, error) {
	var buf bytes.Buffer

	writer := csv.NewWriter(&buf)

	header := []string{
		"event_id",
		"type",
		"occurred_at",
		"account_id",
		"amount_value",
		"amount_currency",
		"status",
		"attempts",
		"processed_at",
		"failed_at",
		"error_code",
		"error_message",
		"batch_id",
		"source",
	}

	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("write CSV header: %w", err)
	}

	for _, evt := range events {
		row := []string{
			evt.EventID,
			string(evt.Type),
			evt.OccurredAt,
			evt.AccountID,
			evt.Amount.Value,
			evt.Amount.Currency,
			string(evt.Status),
			fmt.Sprintf("%d", evt.Attempts),
			evt.ProcessedAt,
			evt.FailedAt,
			evt.LastErrorCode,
			evt.LastErrorMsg,
			evt.BatchID,
			string(evt.Source),
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("write CSV row: %w", err)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV: %w", err)
	}

	return buf.Bytes(), nil
}

func countByStatus(events []*models.DynamoDBEvent, status constants.EventStatus) int {
	count := 0

	for _, evt := range events {
		if evt.Status == status {
			count++
		}
	}

	return count
}
