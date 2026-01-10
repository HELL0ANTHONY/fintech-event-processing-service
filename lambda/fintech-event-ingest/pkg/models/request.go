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
