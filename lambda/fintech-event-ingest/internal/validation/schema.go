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
