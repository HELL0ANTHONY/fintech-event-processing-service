// Package processor implements the stream processing logic for the Lambda function.
package processor

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/pkg/worker"
)

// StreamProcessor processes events from DynamoDB Streams.
type StreamProcessor interface {
	ProcessBatch(ctx context.Context, records []events.DynamoDBEventRecord) error
}

type processor struct {
	repo         repository.EventRepository
	bizValidator validation.BusinessValidator
	workerPool   *worker.Pool
}

// New creates a new StreamProcessor.
func New(
	r repository.EventRepository,
	v validation.BusinessValidator,
	p *worker.Pool,
) StreamProcessor {
	return &processor{
		repo:         r,
		bizValidator: v,
		workerPool:   p,
	}
}

// ProcessBatch processes a batch of DynamoDB Stream records.
func (p *processor) ProcessBatch(
	ctx context.Context,
	records []events.DynamoDBEventRecord,
) error {
	start := time.Now()

	var wg sync.WaitGroup

	var mu sync.Mutex

	var processed, failed int

	for i := range records {
		record := records[i]
		if record.EventName != "INSERT" {
			continue
		}

		wg.Add(1)

		r := record

		p.workerPool.Submit(func() {
			defer wg.Done()

			err := p.processRecord(ctx, &r)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				failed++
			} else {
				processed++
			}
		})
	}

	wg.Wait()

	logger.Info(ctx, "stream batch completed",
		"processed", processed,
		"failed", failed,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return nil
}

func (p *processor) processRecord(
	ctx context.Context,
	record *events.DynamoDBEventRecord,
) error {
	start := time.Now()

	event, err := unmarshalStreamImage(record.Change.NewImage)
	if err != nil {
		logger.Error(ctx, "failed to unmarshal stream record", err)

		return err
	}

	ctx = logger.WithEventContext(ctx, event.EventID, event.BatchID)

	if event.Status != constants.StatusReceived {
		logger.Debug(ctx, "skipping non-RECEIVED event", "status", event.Status)

		return nil
	}

	if err := p.bizValidator.ValidateBusinessRules(event); err != nil {
		return p.handleValidationError(ctx, event, err)
	}

	event.MarkProcessed()

	if err := p.repo.UpdateEventStatus(ctx, event); err != nil {
		return p.handleTransientError(ctx, event, err)
	}

	logger.EventProcessed(ctx, event.EventID, time.Since(start).Milliseconds())

	return nil
}

func (p *processor) handleValidationError(
	ctx context.Context,
	event *models.DynamoDBEvent,
	err error,
) error {
	appErr := toAppError(err)

	event.MarkFailed(appErr.Code, appErr.Message)

	if updateErr := p.repo.UpdateEventStatus(ctx, event); updateErr != nil {
		logger.Error(ctx, "failed to mark event as FAILED", updateErr)
		return updateErr
	}

	logger.EventFailed(ctx, event.EventID, string(appErr.Code), event.Attempts, false)

	return nil
}

func toAppError(err error) *customerrors.AppError {
	var appErr *customerrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return customerrors.Internal(err)
}

func (p *processor) handleTransientError(
	ctx context.Context,
	event *models.DynamoDBEvent,
	err error,
) error {
	appErr := customerrors.WrapTransient(err)

	if event.CanRetry() {
		nextRetry := time.Now().Add(calculateBackoff(event.Attempts))
		event.MarkRetried(appErr.Code, appErr.Message, nextRetry)

		logger.EventRetried(ctx, event.EventID, string(appErr.Code), event.Attempts)
	} else {
		event.MarkFailed(constants.ErrRetryExhausted, "maximum retry attempts exceeded")

		logger.EventFailed(
			ctx,
			event.EventID,
			string(constants.ErrRetryExhausted),
			event.Attempts,
			false,
		)
	}

	return p.repo.UpdateEventStatus(ctx, event)
}

func calculateBackoff(attempts int) time.Duration {
	base := 200 * time.Millisecond

	return base * time.Duration(1<<attempts)
}

func unmarshalStreamImage(
	image map[string]events.DynamoDBAttributeValue,
) (*models.DynamoDBEvent, error) {
	avMap := make(map[string]types.AttributeValue, len(image))

	for k, v := range image {
		avMap[k] = convertToSDKAttributeValue(v)
	}

	var event models.DynamoDBEvent
	if err := attributevalue.UnmarshalMap(avMap, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func convertToSDKAttributeValue(av events.DynamoDBAttributeValue) types.AttributeValue {
	switch av.DataType() {
	case events.DataTypeMap:
		return convertMap(av)
	case events.DataTypeList:
		return convertList(av)
	default:
		return convertScalar(av)
	}
}

func convertScalar(av events.DynamoDBAttributeValue) types.AttributeValue {
	switch av.DataType() {
	case events.DataTypeString:
		return &types.AttributeValueMemberS{Value: av.String()}
	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{Value: av.Number()}
	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{Value: av.Boolean()}
	case events.DataTypeBinary:
		return &types.AttributeValueMemberB{Value: av.Binary()}
	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{Value: av.IsNull()}
	case events.DataTypeStringSet:
		return &types.AttributeValueMemberSS{Value: av.StringSet()}
	case events.DataTypeNumberSet:
		return &types.AttributeValueMemberNS{Value: av.NumberSet()}
	case events.DataTypeBinarySet:
		return &types.AttributeValueMemberBS{Value: av.BinarySet()}
	default:
		return &types.AttributeValueMemberNULL{Value: true}
	}
}

func convertMap(av events.DynamoDBAttributeValue) types.AttributeValue {
	result := make(map[string]types.AttributeValue, len(av.Map()))

	for k, v := range av.Map() {
		result[k] = convertToSDKAttributeValue(v)
	}

	return &types.AttributeValueMemberM{Value: result}
}

func convertList(av events.DynamoDBAttributeValue) types.AttributeValue {
	values := av.List()
	result := make([]types.AttributeValue, len(values))

	for i, v := range values {
		result[i] = convertToSDKAttributeValue(v)
	}

	return &types.AttributeValueMemberL{Value: result}
}
