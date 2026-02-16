package validation

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/shopspring/decimal"
)

// ValidateBusinessRules performs business-specific validations based on event type.
func (v *validator) ValidateBusinessRules(event *models.DynamoDBEvent) error {
	switch event.Type {
	case constants.EventMovementPosted,
		constants.EventMovementReversed,
		constants.EventMovementAdjusted:
		return v.validateMovement(event)
	case constants.EventCreditSettled, constants.EventCreditReversed:
		return v.validateCredit(event)
	case constants.EventReconciliationResult:
		return v.validateReconciliation(event)
	default:
		return nil
	}
}

// validateMovement validates movement events (debit/credit transactions).
func (v *validator) validateMovement(event *models.DynamoDBEvent) error {
	amount, err := event.Amount.Decimal()
	if err != nil {
		return customerrors.BusinessRule(
			constants.ErrInvalidAmount,
			"invalid amount value for movement",
		)
	}

	if amount.IsZero() {
		return customerrors.BusinessRule(
			constants.ErrInvalidMovement,
			"movement amount cannot be zero",
		)
	}

	if direction, ok := event.Metadata["direction"].(string); ok {
		if err := v.validateDirectionConsistency(direction, amount); err != nil {
			return err
		}
	}

	if channel, ok := event.Metadata["channel"].(string); ok {
		if !isValidChannel(channel) {
			return customerrors.BusinessRule(
				constants.ErrInvalidMovement,
				"invalid channel: "+channel,
			)
		}
	}

	return nil
}

func (v *validator) validateDirectionConsistency(
	direction string,
	amount decimal.Decimal,
) error {
	switch direction {
	case "DEBIT":
		if amount.IsPositive() {
			return customerrors.BusinessRule(
				constants.ErrInvalidMovement,
				"DEBIT movement must have negative amount",
			)
		}
	case "CREDIT":
		if amount.IsNegative() {
			return customerrors.BusinessRule(
				constants.ErrInvalidMovement,
				"CREDIT movement must have positive amount",
			)
		}
	default:
		return customerrors.BusinessRule(
			constants.ErrInvalidMovement,
			"invalid direction: must be DEBIT or CREDIT",
		)
	}

	return nil
}

// validateCredit validates credit settlement events.
func (v *validator) validateCredit(event *models.DynamoDBEvent) error {
	amount, err := event.Amount.Decimal()
	if err != nil {
		return customerrors.BusinessRule(
			constants.ErrInvalidAmount,
			"invalid amount value for credit",
		)
	}

	if !amount.IsPositive() {
		return customerrors.BusinessRule(
			constants.ErrInvalidCredit,
			"credit settlement amount must be positive",
		)
	}

	if event.Type == constants.EventCreditSettled {
		ref, hasRef := event.Metadata["reference"]
		if !hasRef || ref == "" {
			return customerrors.BusinessRule(
				constants.ErrMissingReference,
				"credit settlement requires a reference",
			)
		}
	}

	minAmount := decimal.NewFromFloat(0.01)
	if amount.LessThan(minAmount) {
		return customerrors.BusinessRule(
			constants.ErrInvalidCredit,
			"credit amount must be at least 0.01",
		)
	}

	return nil
}

// validateReconciliation validates reconciliation result events.
func (v *validator) validateReconciliation(event *models.DynamoDBEvent) error {
	matchStatus, ok := event.Metadata["matchStatus"].(string)
	if !ok || matchStatus == "" {
		return customerrors.BusinessRule(
			constants.ErrInvalidMatchStatus,
			"matchStatus is required for reconciliation",
		)
	}

	if matchStatus != "MATCHED" && matchStatus != "MISMATCHED" {
		return customerrors.BusinessRule(
			constants.ErrInvalidMatchStatus,
			"matchStatus must be MATCHED or MISMATCHED",
		)
	}

	if matchStatus == "MISMATCHED" {
		reason, hasReason := event.Metadata["reason"].(string)
		if !hasReason || reason == "" {
			return customerrors.BusinessRule(
				constants.ErrMissingMismatchReason,
				"mismatched reconciliation requires a reason",
			)
		}
	}

	if reconID, ok := event.Metadata["reconId"].(string); ok {
		if len(reconID) < 8 {
			return customerrors.BusinessRule(
				constants.ErrInvalidSchema,
				"reconId must be at least 8 characters",
			)
		}
	}

	return nil
}

func isValidChannel(channel string) bool {
	validChannels := map[string]bool{
		"POS":      true,
		"ATM":      true,
		"TRANSFER": true,
		"ONLINE":   true,
		"MOBILE":   true,
		"BRANCH":   true,
		"API":      true,
	}

	return validChannels[channel]
}
