.PHONY: help build build-all test test-integration clean deploy local-dynamodb run-api seed

BINARY_DIR=bin
LAMBDAS=$(shell find cmd -name main.go -exec dirname {} \;)

help:
	@echo "Available commands:"
	@echo "  make build-all          - Build all Lambda functions"
	@echo "  make build FUNCTION=... - Build specific Lambda function"
	@echo "  make test               - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make deploy             - Deploy all Lambdas to AWS"
	@echo "  make deploy-lambda FUNCTION=... - Deploy specific Lambda"
	@echo "  make local-dynamodb     - Start local DynamoDB"
	@echo "  make run-api            - Run API locally"
	@echo "  make seed               - Seed demo users with credentials"
	@echo "  make clean              - Clean build artifacts"

build-all:
	@echo "Building all Lambda functions..."
	@for lambda in $(LAMBDAS); do \
		echo "Building $$lambda..."; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_DIR)/$$(basename $$lambda)/bootstrap $$lambda/main.go; \
		cd $(BINARY_DIR)/$$(basename $$lambda) && zip -q bootstrap.zip bootstrap && cd ../..; \
	done
	@echo "Build complete!"

build:
	@if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION not specified. Usage: make build FUNCTION=exercises/create"; \
		exit 1; \
	fi
	@echo "Building $(FUNCTION)..."
	@mkdir -p $(BINARY_DIR)/$$(basename $(FUNCTION))
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_DIR)/$$(basename $(FUNCTION))/bootstrap cmd/$(FUNCTION)/main.go
	@cd $(BINARY_DIR)/$$(basename $(FUNCTION)) && zip -q bootstrap.zip bootstrap
	@echo "Build complete: $(BINARY_DIR)/$$(basename $(FUNCTION))/bootstrap.zip"

test:
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./internal/... || true
	@if [ -f coverage.out ]; then \
		if go tool cover 2>&1 | grep -q "Usage of"; then \
			go tool cover -html=coverage.out -o coverage.html; \
			echo "Coverage report: coverage.html"; \
		else \
			echo "Coverage tool not available, skipping HTML report generation"; \
		fi \
	else \
		echo "No coverage data generated (no tests or coverage disabled)"; \
	fi

test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

deploy:
	@echo "Deploying all Lambdas..."
	@cd infrastructure/terraform && terraform apply -auto-approve
	@echo "Deploy complete!"

deploy-lambda:
	@if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION not specified. Usage: make deploy-lambda FUNCTION=exercises/create"; \
		exit 1; \
	fi
	@echo "Deploying $(FUNCTION)..."
	@aws lambda update-function-code \
		--function-name training-platform-$$(basename $(FUNCTION)) \
		--zip-file fileb://$(BINARY_DIR)/$$(basename $(FUNCTION))/bootstrap.zip
	@echo "Deploy complete!"

local-dynamodb:
	@echo "Starting local DynamoDB..."
	@docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
	@echo "DynamoDB running on http://localhost:8000"
	@echo "Create table with: aws dynamodb create-table --cli-input-json file://infrastructure/dynamodb-local.json --endpoint-url http://localhost:8000"

run-local:
	@echo "Starting API locally..."
	@go run ./cmd/local

run-lambda-local:
	@echo "Starting Lambda locally with SAM..."
	@sam local start-api

build-lambda:
	@echo "Building Lambda function..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/bootstrap cmd/lambda/main.go
	@cd bin && zip -q bootstrap.zip bootstrap
	@echo "Lambda built: bin/bootstrap.zip"

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

validate-i18n:
	@echo "Validating i18n files..."
	@node scripts/validate-i18n.js

seed:
	@echo "Seeding demo users..."
	@go run cmd/seed/main.go
