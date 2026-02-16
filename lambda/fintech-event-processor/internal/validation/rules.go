package validation

import (
	"slices"
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
)

// ValidateProcessingWindow checks if the event is within the processing window.
func ValidateProcessingWindow(event *models.DynamoDBEvent) error {
	occurredAt, err := time.Parse(time.RFC3339, event.OccurredAt)
	if err != nil {
		return customerrors.BusinessRule(
			constants.ErrInvalidTimestamp,
			"invalid occurred_at timestamp",
		)
	}

	cutoff := time.Now().AddDate(0, 0, -constants.ProcessingWindowDays)
	if occurredAt.Before(cutoff) {
		return customerrors.BusinessRule(
			constants.ErrInvalidTimestamp,
			"event is outside the 90-day processing window",
		)
	}

	return nil
}

// ValidateCurrencyForRegion checks if the currency is allowed for the account's region.
func ValidateCurrencyForRegion(
	event *models.DynamoDBEvent,
	allowedCurrencies []constants.Currency,
) error {
	currency := constants.Currency(event.Amount.Currency)

	if slices.Contains(allowedCurrencies, currency) {
		return nil
	}

	return customerrors.BusinessRule(
		constants.ErrInvalidCurrency,
		"currency not allowed for this region",
	)
}
