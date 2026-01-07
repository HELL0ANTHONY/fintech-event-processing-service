// Package main implements the AWS Lambda function entry point for the fintech event ingest service.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/handler"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

func main() {
	logger.Init()

	n := normalization.New()
	p := processor.New(n)
	h := handler.New(p)

	lambda.Start(h.Handle)
}
