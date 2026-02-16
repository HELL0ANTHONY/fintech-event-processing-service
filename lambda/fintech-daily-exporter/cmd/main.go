// Package main implements the AWS Lambda function entry point for the fintech daily exporter service.
package main

import (
	"context"
	"os"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-daily-exporter/internal/exporter"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-daily-exporter/internal/handler"
	s3repo "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-daily-exporter/internal/repository"
)

func main() {
	ctx := context.Background()

	logger.InitDefault()

	eventRepo, err := repository.NewDynamoDBRepository(ctx)
	if err != nil {
		logger.Error(ctx, "failed to initialize event repository", err)
		os.Exit(1)
	}

	s3Repo, err := s3repo.NewS3Repository(ctx)
	if err != nil {
		logger.Error(ctx, "failed to initialize S3 repository", err)
		os.Exit(1)
	}

	exp := exporter.New(eventRepo, s3Repo)
	h := handler.New(exp)

	logger.Info(ctx, "lambda initialized successfully")

	lambda.Start(h.Handle)
}
