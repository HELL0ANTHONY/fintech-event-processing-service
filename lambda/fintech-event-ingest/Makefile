all: cover

LAMBDA_NAME := $(shell grep -A3 'artifactId:' lambda.yaml | tail -n1 | awk '{ print $$2}')
BINARY_NAME=bootstrap
ARCH=arm64
BUILD_FOLDER=target

hash:
	cd ${BUILD_FOLDER}; \
	openssl dgst -sha256 -binary ./${LAMBDA_NAME}.zip | openssl enc -base64 > ./${LAMBDA_NAME}.hash

zip:
	cd ${BUILD_FOLDER}; \
	zip ${LAMBDA_NAME}.zip ${BINARY_NAME}

build:
	mkdir -p ${BUILD_FOLDER}; \
	cd ${BUILD_FOLDER}; \
	GOOS=linux GOARCH=${ARCH} CGO_ENABLED=0 go build -o ${BINARY_NAME} ../cmd/main.go

update:build zip
	cd ${BUILD_FOLDER}; \
	aws lambda update-function-code --function-name ${LAMBDA_NAME} --zip-file fileb://${LAMBDA_NAME}.zip --region us-east-1 --profile ualintec_dev

lint-prepare:
	 brew list golangci-lint || brew install golangci-lint

lint:lint-prepare
	@golangci-lint run

clean:
	@rm -fr **/*.{out,xml,html}

clean-cache:
	@go clean -cache
	@go clean -testcache
	@go clean -modcache

dependencies:
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -u github.com/onsi/gomega/...

deps:dependencies
	@go mod tidy
	
cover:deps
	@go test ./... -coverprofile=c.out.tmp -coverpkg=./... && cat c.out.tmp | grep -v "_mock.go" > c.out
	
report:cover
	@go tool cover -func c.out | grep "total"
	
html-report:cover
	@go tool cover -html c.out
	
test:deps
	@go test ./...

.PHONY: cover
