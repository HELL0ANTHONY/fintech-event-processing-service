package models

import (
	"github.com/shopspring/decimal"
)

// Amount represents a monetary value with currency.
type Amount struct {
	Value    string `json:"value"    dynamodbav:"value"`
	Currency string `json:"currency" dynamodbav:"currency"`
}

// IsZero returns true if the amount value is zero or empty.
func (a Amount) IsZero() bool {
	if a.Value == "" {
		return true
	}

	d, err := decimal.NewFromString(a.Value)
	if err != nil {
		return true
	}

	return d.IsZero()
}

// IsPositive returns true if the amount value is greater than zero.
func (a Amount) IsPositive() bool {
	d, err := decimal.NewFromString(a.Value)
	if err != nil {
		return false
	}

	return d.IsPositive()
}

// IsNegative returns true if the amount value is less than zero.
func (a Amount) IsNegative() bool {
	d, err := decimal.NewFromString(a.Value)
	if err != nil {
		return false
	}

	return d.IsNegative()
}

// Decimal returns the amount as a decimal.Decimal.
func (a Amount) Decimal() (decimal.Decimal, error) {
	return decimal.NewFromString(a.Value)
}
