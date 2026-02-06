// Package main implements the AWS Lambda function entry point for the fintech event ingest service.
package main

import (
	"context"
	"os"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/handler"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/validation"
)

func main() {
	ctx := context.Background()

	logger.InitDefault()

	repo, err := repository.NewDynamoDBRepository(ctx)
	if err != nil {
		logger.Error(ctx, "failed to initialize repository", err)
		os.Exit(1)
	}

	norm := normalization.New()
	val := validation.New()
	proc := processor.New(norm, val, repo)
	h := handler.New(proc)

	logger.Info(ctx, "lambda initialized successfully")

	lambda.Start(h.Handle)
}
