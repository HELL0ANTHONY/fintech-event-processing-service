// Package validation provides schema validation for incoming events.
package validation

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/shopspring/decimal"
)

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,64}$`)

// ValidateSchema validates a single event against the schema requirements.
// These are "cheap" deterministic validations to accept the event as RECEIVED.
func (v *validator) ValidateSchema(event *models.Event) error {
	if err := v.validateEventID(event.EventID); err != nil {
		return err
	}

	if err := v.validateEventType(event.Type); err != nil {
		return err
	}

	if err := v.validateOccurredAt(event.OccurredAt); err != nil {
		return err
	}

	if err := v.validateAccountID(event.AccountID); err != nil {
		return err
	}

	eventType := constants.EventType(event.Type)
	if eventType.RequiresAmount() {
		if err := v.validateAmount(&event.Amount); err != nil {
			return err
		}
	}

	if err := v.validateMetadata(event.Metadata); err != nil {
		return err
	}

	return nil
}

func (v *validator) validateEventID(eventID string) error {
	if eventID == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"event_id is required",
		)
	}

	if !idRegex.MatchString(eventID) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidSchema,
			"invalid event_id format: must be 8-64 alphanumeric characters with _ or -",
		)
	}

	return nil
}

func (v *validator) validateEventType(eventType string) error {
	if eventType == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"type is required",
		)
	}

	if !constants.IsSupportedEventType(constants.EventType(eventType)) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidEventType,
			"unsupported event type: "+eventType,
		)
	}

	return nil
}

func (v *validator) validateOccurredAt(occurredAt string) error {
	if occurredAt == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"occurred_at is required",
		)
	}

	parsedTime, err := time.Parse(time.RFC3339, occurredAt)
	if err != nil {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidTimestamp,
			"invalid occurred_at: must be RFC3339 format",
		)
	}

	now := time.Now().UTC()

	if parsedTime.After(now.Add(constants.MaxTimestampFuture)) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidTimestamp,
			"occurred_at cannot be more than 10 minutes in the future",
		)
	}

	if parsedTime.Before(now.Add(-constants.MaxTimestampPast)) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidTimestamp,
			"occurred_at cannot be older than 1 year",
		)
	}

	return nil
}

func (v *validator) validateAccountID(accountID string) error {
	if accountID == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"account_id is required",
		)
	}

	return nil
}

func (v *validator) validateAmount(amount *models.Amount) error {
	if amount.Value == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"amount.value is required for this event type",
		)
	}

	if _, err := decimal.NewFromString(amount.Value); err != nil {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidAmount,
			"amount.value must be a valid decimal number",
		)
	}

	if amount.Currency == "" {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"amount.currency is required",
		)
	}

	if !constants.IsSupportedCurrency(constants.Currency(amount.Currency)) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidCurrency,
			"unsupported currency: "+amount.Currency,
		)
	}

	return nil
}

func (v *validator) validateMetadata(metadata map[string]any) error {
	if metadata == nil {
		return nil
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidSchema,
			"invalid metadata format",
		)
	}

	if len(raw) > constants.MaxMetadataSize {
		return customerrors.ValidationWithMessage(
			constants.ErrPayloadTooLarge,
			"metadata exceeds 4KB limit",
		)
	}

	return nil
}

// ValidateRequest validates the entire request before processing individual events.
func ValidateRequest(req *models.Request) error {
	if len(req.Events) == 0 {
		return customerrors.ValidationWithMessage(
			constants.ErrMissingField,
			"at least one event is required",
		)
	}

	if len(req.Events) > constants.MaxEventsPerRequest {
		return customerrors.ValidationWithMessage(
			constants.ErrBatchTooLarge,
			"maximum 500 events per request",
		)
	}

	if req.Source != "" && !constants.IsValidSource(constants.Source(req.Source)) {
		return customerrors.ValidationWithMessage(
			constants.ErrInvalidSchema,
			"invalid source",
		)
	}

	return nil
}
