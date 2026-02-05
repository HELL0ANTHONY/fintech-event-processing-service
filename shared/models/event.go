package models

import (
	"encoding/json"
	"errors"
	"strings"
)

// Event represents a fintech event with its metadata and transaction details.
type Event struct {
	Metadata   map[string]any `json:"metadata,omitempty"`
	Amount     Amount         `json:"amount"`
	EventID    string         `json:"event_id"`
	Type       string         `json:"type"`
	OccurredAt string         `json:"occurred_at"`
	AccountID  string         `json:"account_id"`
}

// Request represents an incoming API request containing fintech events.
type Request struct {
	BatchID string  `json:"batch_id,omitempty"`
	Source  string  `json:"source,omitempty"`
	Events  []Event `json:"events"`
}

// ParseRequest parses a JSON string into a Request struct.
func ParseRequest(body string) (*Request, error) {
	if strings.TrimSpace(body) == "" {
		return nil, errors.New("empty request body")
	}

	var req Request
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return nil, err
	}

	return &req, nil
}
