# Lambda fintech-event-ingest

---

## go.mod

```mod
module github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest

go 1.25.5

require (
	github.com/aws/aws-lambda-go v1.51.1
	github.com/google/uuid v1.6.0
	github.com/shopspring/decimal v1.4.0
)
```

## app_error.go

```go
// Package customerrors defines a structured application error type for consistent error handling.
package customerrors

import (
	"net/http"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
)

// AppError represents a structured application error with additional context.
type AppError struct {
	Cause      error
	Code       constants.ErrorCode
	HTTPStatus int
	Retryable  bool
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}

	return string(e.Code)
}

// Validation creates a new AppError for validation errors.
func Validation(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorValidation,
		Cause:      err,
		Retryable:  false,
		HTTPStatus: http.StatusBadRequest,
	}
}

// DuplicateEvent creates a new AppError for duplicate event ID errors.
func DuplicateEvent() *AppError {
	return &AppError{
		Code:       constants.ErrorDuplicateEvent,
		Retryable:  false,
		HTTPStatus: http.StatusConflict,
		Cause:      nil,
	}
}

// Throttled creates a new AppError for resource throttling errors.
func Throttled(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorResourceThrottled,
		Cause:      err,
		Retryable:  true,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// Internal creates a new AppError for internal server errors.
func Internal(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorInternal,
		Cause:      err,
		Retryable:  false,
		HTTPStatus: http.StatusInternalServerError,
	}
}

```

## sources.go

```go
package constants

// Source represents the origin of the event ingestion request.
type Source string

// Represents the different valid sources for event ingestion.
const (
	SourceAPI         Source = "api"
	SourceBatch       Source = "batch"
	SourceEventBridge Source = "eventbridge"
	SourceReplay      Source = "replay"
	SourceTest        Source = "test"
)

// IsValidSource checks if the provided source is a valid Source.
func IsValidSource(source Source) bool {
	switch source {
	case SourceAPI, SourceBatch, SourceEventBridge, SourceReplay, SourceTest:
		return true
	default:
		return false
	}
}
```

## headers.go

```go
package constants

// Headers returns a map of standard HTTP headers for responses.
func Headers() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Access-Control-Allow-Origin, Access-Control-Allow-Methods, Content-Type",
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Methods": "POST, OPTIONS",
	}
}
```

## events.go

```go
// Package constants contains constant definitions used across the fintech event ingest service.
package constants

// EventStatus defines the possible statuses for an event.
type EventStatus string

// MaxEventsPerRequest defines the maximum number of events allowed in a single request.
var MaxEventsPerRequest = 500

// Predefined event statuses.
const (
	StatusFailed    EventStatus = "FAILED"
	StatusProcessed EventStatus = "PROCESSED"
	StatusReceived  EventStatus = "RECEIVED"
	StatusRetried   EventStatus = "RETRIED"
)

// EventType defines the types of events that can be processed.
type EventType string

// Predefined event types.
const (
	EventMovementPosted   EventType = "movement_posted"
	EventMovementReversed EventType = "movement_reversed"
	EventMovementAdjusted EventType = "movement_adjusted"

	EventCreditSettled  EventType = "credit_settled"
	EventCreditReversed EventType = "credit_reversed"

	EventReconciliationResult EventType = "reconciliation_result"

	EventBatchReceived  EventType = "batch_received"
	EventBatchProcessed EventType = "batch_processed"
	EventBatchFailed    EventType = "batch_failed"
)

// IsSupportedEventType checks if the given event type is supported.
func IsSupportedEventType(et EventType) bool {
	switch et {
	case EventMovementPosted,
		EventMovementReversed,
		EventMovementAdjusted,
		EventCreditSettled,
		EventCreditReversed,
		EventReconciliationResult,
		EventBatchReceived,
		EventBatchProcessed,
		EventBatchFailed:
		return true
	default:
		return false
	}
}
```

## errors.go

```go
package constants

// ErrorCode defines the types of error codes used in the fintech event ingest service.
type ErrorCode string

// Predefined error codes for various error scenarios.
const (
	ErrorValidation     ErrorCode = "validation_error"
	ErrorDuplicateEvent ErrorCode = "duplicate_event_id"
	ErrorConflict       ErrorCode = "conflict"

	ErrorTimeout           ErrorCode = "timeout"
	ErrorResourceThrottled ErrorCode = "resource_throttled"
	ErrorExportFailed      ErrorCode = "export_failed"

	ErrorInternal ErrorCode = "internal_error"
)
```

## error_messages.go

```go
package constants

var errorMessages = map[ErrorCode]string{
	ErrorConflict:          "The request conflicts with the current resource state",
	ErrorDuplicateEvent:    "Event already processed or currently being processed",
	ErrorExportFailed:      "Failed to complete export operation",
	ErrorInternal:          "Unexpected internal error occurred",
	ErrorResourceThrottled: "Service temporarily unavailable due to high load",
	ErrorTimeout:           "Processing exceeded the allowed time limit",
	ErrorValidation:        "The request payload is invalid or violates business rules",
}

// ErrorMessage returns the error message corresponding to the given ErrorCode.
func ErrorMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}

	return "Unknown error"
}

```

## currencies.go

```go
package constants

// Currency defines the type for supported currencies.
type Currency string

// Predefined supported currencies.
const (
	CurrencyUSD Currency = "USD"
	CurrencyARS Currency = "ARS"
)

// IsSupportedCurrency checks if the given currency is supported.
func IsSupportedCurrency(c Currency) bool {
	switch c {
	case CurrencyUSD, CurrencyARS:
		return true
	default:
		return false
	}
}
```

## correlation.go

```go
// Package httpx provides HTTP-related utilities for AWS Lambda functions.
package httpx

import "github.com/aws/aws-lambda-go/events"

// CorrelationID extracts the correlation ID from the request headers.
func CorrelationID(req *events.APIGatewayProxyRequest) string {
	if id := req.Headers["X-Correlation-Id"]; id != "" {
		return id
	}

	return req.RequestContext.RequestID
}

```

## normalization.go

```go
// Package normalization provides interfaces and implementations for data normalization.
package normalization

import (
	"github.com/google/uuid"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// Normalizable defines the contract for normalizing data.
type Normalizable interface {
	Normalize(req *models.Request) *models.Request
}

// Normalizer implements the Normalizable interface.
type Normalizer struct{}

// New creates a new instance of Normalizer.
func New() Normalizable {
	return &Normalizer{}
}

// Normalize processes the input request and ensures required fields are set.
func (n *Normalizer) Normalize(req *models.Request) *models.Request {
	if req.Source == "" {
		req.Source = "api"
	}

	if req.BatchID == "" {
		req.BatchID = uuid.New().String()
	}

	return req
}

```

## handler.go

```go
// Package handler implements the AWS Lambda handler for processing fintech events.
package handler

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

type Handler struct {
	p processor.RequestProcessor
}

// New creates a new Handler with the given RequestProcessor.
func New(p processor.RequestProcessor) Handler {
	return Handler{
		p: p,
	}
}

// Handle is the AWS Lambda handler function.
func (h Handler) Handle(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	ctx = logger.With(
		ctx,
		slog.String("request_id", req.RequestContext.RequestID),
		slog.String("path", req.Path),
	)

	return h.p.Process(ctx, req)
}

```

## main.go

```go
// Package main implements the AWS Lambda function entry point for the fintech event ingest service.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/handler"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

func main() {
	logger.Init()

	n := normalization.New()
	v := validation.New()
	p := processor.New(n, v)

	h := handler.New(p)

	lambda.Start(h.Handle)
}

```

## request.go

```go
// Package models this file defines the data structures for handling incoming fintech event requests.
package models

import (
	"encoding/json"
	"errors"
)

// Amount represents the monetary value and currency of a transaction.
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// Event represents a fintech event with its associated metadata and details.
type Event struct {
	Metadata   map[string]any `json:"metadata"`
	Amount     Amount         `json:"amount"`
	EventID    string         `json:"event_id"`
	Type       string         `json:"type"`
	OccurredAt string         `json:"occurred_at"`
	AccountID  string         `json:"account_id"`
}

// Request represents an incoming request containing a fintech event.
type Request struct {
	Source  string  `json:"source,omitempty"`
	BatchID string  `json:"batch_id"`
	Event   []Event `json:"event"`
}

// NewJSONRequest parses a JSON string into a slice of Request structs.
func NewJSONRequest(rawBody string) (*Request, error) {
	if rawBody == "" {
		return nil, errors.New("empty request body")
	}

	var requests *Request

	if err := json.Unmarshal([]byte(rawBody), &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

```

## dynamoDB.go

```go
package models

import "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"

// DynamoDBEvent represents the structure of an event stored in DynamoDB.
type DynamoDBEvent struct {
	Metadata      map[string]any        `json:"metadata"                  dynamodbav:"metadata"`
	Amount        Amount                `json:"amount"                    dynamodbav:"amount"`
	Status        constants.EventStatus `json:"status"                    dynamodbav:"status"`
	ReceivedAt    string                `json:"received_at"               dynamodbav:"received_at"`
	ProcessedAt   string                `json:"processed_at,omitempty"    dynamodbav:"processed_at,omitempty"`
	FailedAt      string                `json:"failed_at,omitempty"       dynamodbav:"failed_at,omitempty"`
	LastErrorCode string                `json:"last_error_code,omitempty" dynamodbav:"last_error_code,omitempty"`
	NextRetryAt   string                `json:"next_retry_at,omitempty"   dynamodbav:"next_retry_at,omitempty"`
	BatchID       string                `json:"batch_id"                  dynamodbav:"batch_id"`
	EventID       string                `json:"event_id"                  dynamodbav:"event_id"`
	Type          constants.EventType   `json:"type"                      dynamodbav:"type"`
	OccurredAt    string                `json:"occurred_at"               dynamodbav:"occurred_at"`
	AccountID     string                `json:"account_id"                dynamodbav:"account_id"`
	Source        constants.Source      `json:"source"                    dynamodbav:"source"`
	Attempts      int                   `json:"attempts"                  dynamodbav:"attempts"`
}

// NewDynamoDBEventFromRequest creates a new DynamoDBEvent from a Request.
func NewDynamoDBEventFromRequest(
	event *Event,
	source constants.Source,
	receivedAt, batchID string,
) *DynamoDBEvent {
	return &DynamoDBEvent{
		Source:        source,
		BatchID:       batchID,
		ReceivedAt:    receivedAt,
		Status:        constants.StatusReceived,
		Attempts:      0,
		Metadata:      event.Metadata,
		Amount:        event.Amount,
		EventID:       event.EventID,
		Type:          constants.EventType(event.Type),
		OccurredAt:    event.OccurredAt,
		AccountID:     event.AccountID,
		ProcessedAt:   "",
		FailedAt:      "",
		LastErrorCode: "",
		NextRetryAt:   "",
	}
}

```

## logger.go

```go
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

const (
	KeyService  = "service"
	KeyFunction = "fn"
	KeyError    = "error"
)

// Init initializes the global logger. Call once in main.
func Init() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		level = slog.LevelDebug
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			return a
		},
	})

	base := slog.New(h)
	if svc := os.Getenv("SERVICE_NAME"); svc != "" {
		base = base.With(slog.String(KeyService, svc))
	}

	slog.SetDefault(base)
	log.SetFlags(0)
	log.SetOutput(&writer{base})
}

type writer struct {
	logger *slog.Logger
}

func (w *writer) Write(p []byte) (int, error) {
	w.logger.Info(strings.TrimSuffix(string(p), "\n"))
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
	from(ctx).InfoContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// Error logs at ERROR level with automatic caller detection.
func Error(ctx context.Context, msg string, err error, args ...any) {
	from(
		ctx,
	).ErrorContext(ctx, msg, append(args, slog.String(KeyFunction, caller()), slog.String(KeyError, err.Error()))...)
}

// Debug logs at DEBUG level with automatic caller detection.
func Debug(ctx context.Context, msg string, args ...any) {
	from(ctx).DebugContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// Warn logs at WARN level with automatic caller detection.
func Warn(ctx context.Context, msg string, args ...any) {
	from(ctx).WarnContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

// ErrorAttrs logs at ERROR level with automatic caller detection and additional attributes.
func ErrorAttrs(ctx context.Context, msg string, args ...any) {
	from(ctx).ErrorContext(ctx, msg, append(args, slog.String(KeyFunction, caller()))...)
}

```

## errors_response.go

```go
package httpx

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	customerrors "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/errors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

type errorDetail struct {
	Code    constants.ErrorCode `json:"code"`
	Message string              `json:"message"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

const fallbackErrorJSON = `{"error":{"code":"internal_error","message":"Internal server error"}}`

// WithCorrelationID adds correlation ID to existing headers.
func WithCorrelationID(headers map[string]string, cid string) map[string]string {
	headers["X-Correlation-Id"] = cid
	return headers
}

// FromAppError constructs an APIGatewayProxyResponse from an AppError.
func FromAppError(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
	appErr *customerrors.AppError,
) events.APIGatewayProxyResponse {
	cid := CorrelationID(req)
	msg := constants.ErrorMessage(appErr.Code)

	attrs := []any{
		slog.String("correlation_id", cid),
		slog.String("error_code", string(appErr.Code)),
		slog.Bool("retryable", appErr.Retryable),
		slog.String("message", msg),
	}

	if appErr.Cause != nil {
		attrs = append(attrs, slog.String("cause", appErr.Cause.Error()))
	}

	logger.ErrorAttrs(ctx, "request failed", attrs...)

	body, err := json.Marshal(errorBody{
		Error: errorDetail{
			Code:    appErr.Code,
			Message: msg,
		},
	})
	if err != nil {
		logger.Error(ctx, "failed to marshal error response", err)

		body = []byte(fallbackErrorJSON)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:        appErr.HTTPStatus,
		Body:              string(body),
		Headers:           WithCorrelationID(constants.Headers(), cid),
		MultiValueHeaders: map[string][]string{},
		IsBase64Encoded:   false,
	}
}
```

## processor.go

```go
// Package processor implements the request processing logic for the Lambda function.
package processor

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	customerrors "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/errors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/httpx"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// RequestProcessor defines the contract for processing incoming requests.
type RequestProcessor interface {
	Process(
		ctx context.Context,
		req *events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error)
}

// Processor implements the RequestProcessor interface.
type Processor struct {
	n normalization.Normalizable
	v validation.Validatable
}

// New creates a new RequestProcessor implementation.
func New(n normalization.Normalizable, v validation.Validatable) RequestProcessor {
	return &Processor{
		n: n,
		v: v,
	}
}

// Process handles the business logic for the incoming request.
func (p *Processor) Process(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	transactions, err := models.NewJSONRequest(req.Body)
	if err != nil {
		appErr := customerrors.Validation(err)

		return httpx.FromAppError(ctx, req, appErr), nil
	}

	normalizedRequest := p.n.Normalize(transactions)

	// NOTE: No imprimir datos demasiado grandes en logs reales.
	logger.Debug(
		ctx,
		"normalized request",
		slog.Any("payload", normalizedRequest),
	)

	// NOTE: Valida los schemas y luego se guardan en la base de datos con el status RECEIVED.
	if err := p.v.ValidateSchema(normalizedRequest); err != nil {
		appErr := customerrors.Validation(err)

		return httpx.FromAppError(ctx, req, appErr), nil
	}

	logger.Info(
		ctx,
		"request processed successfully",
		slog.Int("transaction_count", len(normalizedRequest.Event)),
	)

	return events.APIGatewayProxyResponse{
		StatusCode:        http.StatusOK,
		Body:              "Request processed successfully",
		IsBase64Encoded:   false,
		Headers:           constants.Headers(),
		MultiValueHeaders: map[string][]string{},
	}, nil
}
```

## repository.go

```go
package repository
```

## business.go

```go
package validation

import (
	"log"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// ValidateBusinessRules performs business rule validations on the request data.
func (v *Validator) ValidateBusinessRules(data *models.Request) error {
	log.Printf("Validating business rules for request ID: %#v", data)

	return nil
}
```

## schema.go

```go
package validation

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/shopspring/decimal"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

const maxMetadataSize = 4 * 1024

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,64}$`)

// ValidateSchema defines the contract for validating request schemas.
func (v *Validator) ValidateSchema(data *models.Request) error {
	if len(data.Event) == 0 {
		return errors.New("no events provided")
	}

	if len(data.Event) > constants.MaxEventsPerRequest {
		return errors.New("too many events in request")
	}

	if !constants.IsValidSource(constants.Source(data.Source)) {
		return errors.New("invalid source")
	}

	if !idRegex.MatchString(data.BatchID) {
		return errors.New("invalid batch ID")
	}

	for i := range data.Event {
		if err := v.validateEvent(&data.Event[i]); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateEvent(event *models.Event) error {
	if !idRegex.MatchString(event.EventID) {
		return errors.New("invalid event ID")
	}

	if !constants.IsSupportedEventType(constants.EventType(event.Type)) {
		return errors.New("unsupported event type")
	}

	if _, err := time.Parse(time.RFC3339, event.OccurredAt); err != nil {
		return errors.New("invalid occurredAt timestamp")
	}

	if event.AccountID == "" {
		return errors.New("invalid account ID")
	}

	if err := v.validateAmount(&event.Amount); err != nil {
		return err
	}

	if err := v.validateMetadata(event.Metadata); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateAmount(amount *models.Amount) error {
	if amount.Value == "" {
		return nil
	}

	if _, err := decimal.NewFromString(amount.Value); err != nil {
		return errors.New("invalid amount value")
	}

	if !constants.IsSupportedCurrency(constants.Currency(amount.Currency)) {
		return errors.New("invalid currency")
	}

	return nil
}

func (v *Validator) validateMetadata(metadata map[string]any) error {
	if metadata == nil {
		return nil
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		return errors.New("invalid metadata format")
	}

	if len(raw) > maxMetadataSize {
		return errors.New("metadata exceeds size limit")
	}

	return nil
}
```

## validation.go

```go
// Package validation provides interfaces and implementations for validating request schemas and business rules.
package validation

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// Validatable defines the contract for validating request schemas.
type Validatable interface {
	ValidateSchema(data *models.Request) error
	ValidateBusinessRules(data *models.Request) error
}

// Validator implements the Validatable interface.
type Validator struct{}

// New creates a new instance of Validator.
func New() Validatable {
	return &Validator{}
}
```

## processor.go

```go
// Package processor implements the request processing logic for the Lambda function.
package processor

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	customerrors "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/errors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/httpx"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// RequestProcessor defines the contract for processing incoming requests.
type RequestProcessor interface {
	Process(
		ctx context.Context,
		req *events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error)
}

// Processor implements the RequestProcessor interface.
type Processor struct {
	n normalization.Normalizable
	v validation.Validatable
}

// New creates a new RequestProcessor implementation.
func New(n normalization.Normalizable, v validation.Validatable) RequestProcessor {
	return &Processor{
		n: n,
		v: v,
	}
}

// Process handles the business logic for the incoming request.
func (p *Processor) Process(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	transactions, err := models.NewJSONRequest(req.Body)
	if err != nil {
		appErr := customerrors.Validation(err)

		return httpx.FromAppError(ctx, req, appErr), nil
	}

	normalizedRequest := p.n.Normalize(transactions)

	// NOTE: No imprimir datos demasiado grandes en logs reales.
	logger.Debug(
		ctx,
		"normalized request",
		slog.Any("payload", normalizedRequest),
	)

	// NOTE: Valida los schemas y luego se guardan en la base de datos con el status RECEIVED.
	if err := p.v.ValidateSchema(normalizedRequest); err != nil {
		appErr := customerrors.Validation(err)

		return httpx.FromAppError(ctx, req, appErr), nil
	}

	logger.Info(
		ctx,
		"request processed successfully",
		slog.Int("transaction_count", len(normalizedRequest.Event)),
	)

	return events.APIGatewayProxyResponse{
		StatusCode:        http.StatusOK,
		Body:              "Request processed successfully",
		IsBase64Encoded:   false,
		Headers:           constants.Headers(),
		MultiValueHeaders: map[string][]string{},
	}, nil
}
```
