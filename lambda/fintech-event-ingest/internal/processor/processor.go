// Package processor implements the request processing logic for the Lambda function.
package processor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"
	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/httpx"
)

// RequestProcessor defines the contract for processing ingest requests.
type RequestProcessor interface {
	Process(
		ctx context.Context,
		req *events.APIGatewayProxyRequest,
		correlationID string,
	) (events.APIGatewayProxyResponse, error)
}

type Processor struct {
	normalizer normalization.Normalizer
	validator  validation.SchemaValidator
	repo       repository.EventRepository
}

// New creates a new RequestProcessor.
func New(
	n normalization.Normalizer,
	v validation.SchemaValidator,
	r repository.EventRepository,
) RequestProcessor {
	return &Processor{
		normalizer: n,
		validator:  v,
		repo:       r,
	}
}

// Process handles the incoming request: parse -> normalize -> validate -> persist.
func (p *Processor) Process(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
	correlationID string,
) (events.APIGatewayProxyResponse, error) {
	start := time.Now()

	request, err := models.ParseRequest(req.Body)
	if err != nil {
		appErr := customerrors.Validation(constants.ErrInvalidSchema, err)

		return httpx.ErrorResponse(appErr, correlationID), nil
	}

	if err := validation.ValidateRequest(request); err != nil {
		return httpx.ErrorResponse(toAppError(err), correlationID), nil
	}

	request = p.normalizer.Normalize(request)
	ctx = logger.WithEventContext(ctx, "", request.BatchID)

	logger.BatchReceived(ctx, request.BatchID, request.Source, len(request.Events))

	validEvents, rejectedResults := p.validateEvents(ctx, request)

	if len(validEvents) == 0 && len(rejectedResults) > 0 {
		return p.buildResponse(ctx, request.BatchID, nil, rejectedResults, correlationID), nil
	}

	persistResults := p.persistEvents(ctx, validEvents, request.Source, request.BatchID)

	allResults := mergeResults(rejectedResults, persistResults)

	processed := countByStatus(allResults, "RECEIVED")
	duplicates := countByStatus(allResults, "DUPLICATE")
	rejected := len(rejectedResults)

	logger.BatchCompleted(
		ctx,
		request.BatchID,
		processed,
		rejected+duplicates,
		time.Since(start).Milliseconds(),
	)

	return p.buildResponse(
		ctx,
		request.BatchID,
		persistResults,
		rejectedResults,
		correlationID,
	), nil
}

func (p *Processor) validateEvents(
	ctx context.Context,
	request *models.Request,
) ([]*models.Event, []models.EventResult) {
	var valid []*models.Event

	var rejected []models.EventResult

	for i := range request.Events {
		evt := &request.Events[i]

		if err := p.validator.ValidateSchema(evt); err != nil {
			appErr := toAppError(err)

			rejected = append(rejected, models.EventResult{
				EventID:   evt.EventID,
				Status:    "REJECTED",
				ErrorCode: string(appErr.Code),
				Message:   appErr.Message,
			})

			logger.Debug(ctx, "event validation failed",
				"event_id", evt.EventID,
				"error_code", appErr.Code,
			)

			continue
		}

		valid = append(valid, evt)
	}

	return valid, rejected
}

func (p *Processor) persistEvents(
	ctx context.Context,
	inputEvents []*models.Event,
	source, batchID string,
) []models.EventResult {
	dbEvents := make([]*models.DynamoDBEvent, len(inputEvents))

	for i, evt := range inputEvents {
		dbEvents[i] = models.NewDynamoDBEvent(evt, constants.Source(source), batchID)
	}

	batchResults, err := p.repo.SaveEventsBatch(ctx, dbEvents)
	if err != nil {
		logger.Error(ctx, "batch save failed", err)
	}

	results := make([]models.EventResult, len(batchResults))

	for i, br := range batchResults {
		result := models.EventResult{EventID: br.EventID}

		switch {
		case br.Success:
			result.Status = "RECEIVED"
		case br.Duplicate:
			result.Status = "DUPLICATE"
			result.ErrorCode = string(constants.ErrDuplicateEvent)
			result.Message = constants.ErrDuplicateEvent.Message()
		default:
			result.Status = "FAILED"
			result.ErrorCode = string(customerrors.GetErrorCode(br.Error))

			if br.Error != nil {
				result.Message = br.Error.Error()
			}
		}

		results[i] = result
	}

	return results
}

func (p *Processor) buildResponse(
	ctx context.Context,
	batchID string,
	persistResults, rejectedResults []models.EventResult,
	correlationID string,
) events.APIGatewayProxyResponse {
	allResults := mergeResults(rejectedResults, persistResults)

	response := models.IngestResponse{
		BatchID:   batchID,
		Received:  countByStatus(allResults, "RECEIVED"),
		Rejected:  countByStatus(allResults, "REJECTED") + countByStatus(allResults, "FAILED"),
		Duplicate: countByStatus(allResults, "DUPLICATE"),
		Results:   allResults,
	}

	body, err := json.Marshal(response)
	if err != nil {
		logger.Error(ctx, "failed to marshal response", err)

		return httpx.ErrorResponse(customerrors.Internal(err), correlationID)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Headers:    httpx.WithCorrelationID(constants.StandardHeaders(), correlationID),
	}
}

// toAppError converts an error to AppError using errors.As.
func toAppError(err error) *customerrors.AppError {
	var appErr *customerrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return customerrors.Validation(constants.ErrInvalidSchema, err)
}

// mergeResults combines two slices of EventResult without modifying the originals.
func mergeResults(a, b []models.EventResult) []models.EventResult {
	result := make([]models.EventResult, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)

	return result
}

func countByStatus(results []models.EventResult, status string) int {
	count := 0

	for _, r := range results {
		if r.Status == status {
			count++
		}
	}

	return count
}
