// Package repository provides S3 operations for the exporter service.
package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	appconfig "github.com/HELL0ANTHONY/fintech-event-processing-service/shared/config"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Repository handles S3 operations.
type S3Repository interface {
	Upload(ctx context.Context, key string, data []byte) (*models.S3Info, error)
}

// S3Client abstracts the S3 client for testing.
type S3Client interface {
	PutObject(
		ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
}

type s3Repository struct {
	client S3Client
	bucket string
}

// NewS3Repository creates a new S3 repository.
// The context is used for AWS config loading, enabling trace propagation and timeouts.
func NewS3Repository(ctx context.Context) (S3Repository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	bucket, err := appconfig.GetEnvRequiredNonEmpty(constants.EnvS3Bucket)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "s3 repository initialized", "bucket", bucket)

	return &s3Repository{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// NewS3RepositoryWithClient creates a repository with an injected client.
// Use this for testing or when you need custom client configuration.
func NewS3RepositoryWithClient(client S3Client, bucket string) (S3Repository, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	if bucket == "" {
		return nil, errors.New("bucket name cannot be empty")
	}

	return &s3Repository{
		client: client,
		bucket: bucket,
	}, nil
}

// Upload uploads data to S3 and returns the S3 info.
func (r *s3Repository) Upload(
	ctx context.Context,
	key string,
	data []byte,
) (*models.S3Info, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("text/csv"),
	}

	result, err := r.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	etag := ""
	if result.ETag != nil {
		etag = *result.ETag
	}

	logger.Info(ctx, "file uploaded to S3",
		"bucket", r.bucket,
		"key", key,
		"size_bytes", len(data),
	)

	return &models.S3Info{
		Bucket: r.bucket,
		Key:    key,
		ETag:   etag,
	}, nil
}
