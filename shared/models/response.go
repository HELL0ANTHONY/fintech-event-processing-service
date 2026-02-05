package models

// IngestResponse represents the response for the ingest endpoint.
type IngestResponse struct {
	BatchID   string        `json:"batch_id"`
	Results   []EventResult `json:"results,omitempty"`
	Received  int           `json:"received"`
	Rejected  int           `json:"rejected"`
	Duplicate int           `json:"duplicate"`
}

// EventResult represents the processing result for a single event.
type EventResult struct {
	EventID   string `json:"event_id"`
	Status    string `json:"status"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ExportResponse represents the response for the export Lambda.
type ExportResponse struct {
	S3            *S3Info  `json:"s3,omitempty"`
	ExportDate    string   `json:"export_date"`
	StatusFilter  []string `json:"status_filter"`
	ExportedCount int      `json:"exported_count"`
	FailedCount   int      `json:"failed_count"`
	DurationMs    int64    `json:"duration_ms"`
}

// S3Info contains information about the exported S3 object.
type S3Info struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"e_tag,omitempty"`
}
