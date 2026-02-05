module github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest

go 1.25.6

require (
	github.com/aws/aws-lambda-go v1.51.1
	github.com/google/uuid v1.6.0
	github.com/shopspring/decimal v1.4.0
)

replace github.com/HELL0ANTHONY/fintech-event-processing-service/shared => ../../shared
