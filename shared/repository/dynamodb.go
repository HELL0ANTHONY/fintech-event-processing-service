// Package repository provides data access abstractions for the event processing service.
package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appconfig "github.com/HELL0ANTHONY/fintech-event-processing-service/shared/config"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
)

// DynamoDBClient abstracts the DynamoDB client for testing.
type DynamoDBClient interface {
	PutItem(
		ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.PutItemOutput, error)
	UpdateItem(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)
	GetItem(
		ctx context.Context,
		params *dynamodb.GetItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.GetItemOutput, error)
	Query(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error)
}

// dynamoDBRepository implements EventRepository using DynamoDB.
type dynamoDBRepository struct {
	client    DynamoDBClient
	tableName string
}

// NewDynamoDBRepository creates a new DynamoDB-backed event repository.
// The context is used for AWS config loading, enabling trace propagation and timeouts.
func NewDynamoDBRepository(ctx context.Context) (EventRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	tableName, err := appconfig.GetEnvRequiredNonEmpty(constants.EnvTableName)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "dynamodb repository initialized", "table", tableName)

	return &dynamoDBRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

// NewDynamoDBRepositoryWithClient creates a repository with an injected client.
// Use this for testing or when you need custom client configuration.
func NewDynamoDBRepositoryWithClient(
	client DynamoDBClient,
	tableName string,
) (EventRepository, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	if tableName == "" {
		return nil, errors.New("table name cannot be empty")
	}

	return &dynamoDBRepository{
		client:    client,
		tableName: tableName,
	}, nil
}

// SaveEvent persists a single event with conditional check for idempotency.
func (r *dynamoDBRepository) SaveEvent(ctx context.Context, event *models.DynamoDBEvent) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return customerrors.Internal(fmt.Errorf("marshal event: %w", err))
	}

	condition := expression.AttributeNotExists(expression.Name("PK"))

	expr, err := expression.NewBuilder().WithCondition(condition).Build()
	if err != nil {
		return customerrors.Internal(fmt.Errorf("build expression: %w", err))
	}

	input := &dynamodb.PutItemInput{
		TableName:                 aws.String(r.tableName),
		Item:                      item,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return customerrors.Duplicate(event.EventID)
		}

		return customerrors.WrapTransient(err)
	}

	logger.EventReceived(ctx, event.EventID, string(event.Type), event.AccountID, event.BatchID)

	return nil
}

// SaveEventsBatch persists multiple events using concurrent PutItem operations.
func (r *dynamoDBRepository) SaveEventsBatch(
	ctx context.Context,
	events []*models.DynamoDBEvent,
) ([]BatchResult, error) {
	results := make([]BatchResult, len(events))

	var wg sync.WaitGroup

	var mu sync.Mutex

	const maxConcurrent = 10

	sem := make(chan struct{}, maxConcurrent)

	for i, event := range events {
		wg.Add(1)

		go func(idx int, evt *models.DynamoDBEvent) {
			defer wg.Done()

			sem <- struct{}{}

			defer func() { <-sem }()

			err := r.SaveEvent(ctx, evt)

			mu.Lock()
			defer mu.Unlock()

			results[idx] = BatchResult{
				EventID: evt.EventID,
				Success: err == nil,
				Duplicate: err != nil &&
					customerrors.GetErrorCode(err) == constants.ErrDuplicateEvent,
				Error: err,
			}
		}(i, event)
	}

	wg.Wait()

	return results, nil
}

// UpdateEventStatus updates the status and related fields of an event.
func (r *dynamoDBRepository) UpdateEventStatus(
	ctx context.Context,
	event *models.DynamoDBEvent,
) error {
	update := expression.
		Set(expression.Name("status"), expression.Value(event.Status)).
		Set(expression.Name("status_date"), expression.Value(event.StatusDate)).
		Set(expression.Name("attempts"), expression.Value(event.Attempts))

	switch event.Status {
	case constants.StatusProcessed:
		update = update.Set(expression.Name("processed_at"), expression.Value(event.ProcessedAt))
	case constants.StatusFailed:
		update = update.
			Set(expression.Name("failed_at"), expression.Value(event.FailedAt)).
			Set(expression.Name("last_error_code"), expression.Value(event.LastErrorCode)).
			Set(expression.Name("last_error_msg"), expression.Value(event.LastErrorMsg))
	case constants.StatusRetried:
		update = update.
			Set(expression.Name("last_error_code"), expression.Value(event.LastErrorCode)).
			Set(expression.Name("last_error_msg"), expression.Value(event.LastErrorMsg)).
			Set(expression.Name("next_retry_at"), expression.Value(event.NextRetryAt))
	}

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return customerrors.Internal(fmt.Errorf("build update expression: %w", err))
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: event.EventID},
			"SK": &types.AttributeValueMemberS{Value: event.AccountID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = r.client.UpdateItem(ctx, input)
	if err != nil {
		return customerrors.WrapTransient(err)
	}

	return nil
}

// GetEvent retrieves an event by its ID.
func (r *dynamoDBRepository) GetEvent(
	ctx context.Context,
	eventID, accountID string,
) (*models.DynamoDBEvent, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: eventID},
			"SK": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, customerrors.WrapTransient(err)
	}

	if result.Item == nil {
		return nil, customerrors.NotFound("event", eventID)
	}

	var event models.DynamoDBEvent
	if err := attributevalue.UnmarshalMap(result.Item, &event); err != nil {
		return nil, customerrors.Internal(fmt.Errorf("unmarshal event: %w", err))
	}

	return &event, nil
}

// QueryEventsByStatus retrieves events by status and date for the exporter.
func (r *dynamoDBRepository) QueryEventsByStatus(
	ctx context.Context,
	status string,
	date string,
	limit int32,
) ([]*models.DynamoDBEvent, error) {
	statusDate := status + "#" + date

	keyCond := expression.Key("status_date").Equal(expression.Value(statusDate))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, customerrors.Internal(fmt.Errorf("build query expression: %w", err))
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		IndexName:                 aws.String("status-date-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	var events []*models.DynamoDBEvent

	var lastKey map[string]types.AttributeValue

	for {
		if lastKey != nil {
			input.ExclusiveStartKey = lastKey
		}

		result, err := r.client.Query(ctx, input)
		if err != nil {
			return nil, customerrors.WrapTransient(err)
		}

		for _, item := range result.Items {
			var event models.DynamoDBEvent
			if err := attributevalue.UnmarshalMap(item, &event); err != nil {
				logger.Warn(ctx, "failed to unmarshal event, skipping", "error", err.Error())

				continue
			}

			events = append(events, &event)
		}

		if result.LastEvaluatedKey == nil {
			break
		}

		lastKey = result.LastEvaluatedKey
	}

	return events, nil
}
