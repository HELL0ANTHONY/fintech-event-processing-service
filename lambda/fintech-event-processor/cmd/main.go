// Package main implements the AWS Lambda function entry point for the fintech event processor service.
package main

import (
	"context"
	"os"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/config"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/repository"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/internal/handler"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/internal/validation"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/pkg/worker"
)

const defaultWorkerPoolSize = 5

func main() {
	ctx := context.Background()

	logger.InitDefault()

	repo, err := repository.NewDynamoDBRepository(ctx)
	if err != nil {
		logger.Error(ctx, "failed to initialize repository", err)
		os.Exit(1)
	}

	poolSize := config.GetEnvInt(constants.EnvWorkerPool, defaultWorkerPoolSize)

	bizValidator := validation.New()
	pool := worker.NewPool(poolSize)
	proc := processor.New(repo, bizValidator, pool)
	h := handler.New(proc)

	logger.Info(ctx, "lambda initialized successfully", "worker_pool_size", poolSize)

	lambda.Start(h.Handle)
}
