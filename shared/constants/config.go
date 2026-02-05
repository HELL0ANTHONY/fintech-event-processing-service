// Package constants contains constant definitions used across the fintech event processing service.
package constants

import "time"

// Processing configuration constants.
const (
	MaxRetryAttempts     = 3
	MaxMetadataSize      = 4 * 1024   // 4KB.
	MaxPayloadSize       = 256 * 1024 // 256KB.
	MaxTimestampFuture   = 10 * time.Minute
	MaxTimestampPast     = 365 * 24 * time.Hour // 1 year.
	ProcessingWindowDays = 90
)

// Environment variable names.
const (
	EnvTableName     = "DYNAMODB_TABLE_NAME"
	EnvServiceName   = "SERVICE_NAME"
	EnvLogLevelDebug = "LOG_LEVEL_DEBUG"
	EnvS3Bucket      = "S3_EXPORT_BUCKET"
	EnvAWSRegion     = "AWS_REGION"
	EnvWorkerPool    = "WORKER_POOL_SIZE"
)
