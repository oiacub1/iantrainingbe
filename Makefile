.PHONY: help build build-all test test-integration clean deploy local-dynamodb run-local seed

BINARY_DIR=bin

help:
	@echo "Available commands:"
	@echo "  make build-all          - Build API for all environments"
	@echo "  make build ENV=...      - Build API for specific env (lambda|local)"
	@echo "  make test               - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make deploy             - Deploy API Lambda to AWS"
	@echo "  make local-dynamodb     - Start local DynamoDB"
	@echo "  make run-local          - Run API locally"
	@echo "  make seed               - Seed demo users with credentials"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Architecture: Single unified API codebase"
	@echo "  - cmd/api/main.go runs as Lambda or Local based on environment"

build-all:
	@echo "Building API for all environments..."
	@mkdir -p $(BINARY_DIR)/lambda
	@mkdir -p $(BINARY_DIR)/local
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_DIR)/lambda/bootstrap ./cmd/api
	@cd $(BINARY_DIR)/lambda && zip -q bootstrap.zip bootstrap && cd ../..
	@go build -o $(BINARY_DIR)/local/api ./cmd/api
	@echo "Build complete!"
	@echo "  Lambda: $(BINARY_DIR)/lambda/bootstrap.zip"
	@echo "  Local:  $(BINARY_DIR)/local/api"

build:
	@if [ -z "$(ENV)" ]; then \
		echo "Error: ENV not specified. Usage: make build ENV=lambda|local"; \
		exit 1; \
	fi
	@echo "Building API for $(ENV)..."
	@if [ "$(ENV)" = "lambda" ]; then \
		mkdir -p $(BINARY_DIR)/lambda; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_DIR)/lambda/bootstrap ./cmd/api; \
		cd $(BINARY_DIR)/lambda && zip -q bootstrap.zip bootstrap && cd ../..; \
		echo "Lambda built: $(BINARY_DIR)/lambda/bootstrap.zip"; \
	elif [ "$(ENV)" = "local" ]; then \
		mkdir -p $(BINARY_DIR)/local; \
		go build -o $(BINARY_DIR)/local/api ./cmd/api; \
		echo "Local built: $(BINARY_DIR)/local/api"; \
	else \
		echo "Error: ENV must be 'lambda' or 'local'"; \
		exit 1; \
	fi

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
	@echo "Deploying API Lambda..."
	@if [ ! -f "$(BINARY_DIR)/lambda/bootstrap.zip" ]; then \
		echo "Lambda not built. Run 'make build ENV=lambda' first"; \
		exit 1; \
	fi
	@aws lambda update-function-code \
		--function-name IanTrainingBackend \
		--zip-file fileb://$(BINARY_DIR)/lambda/bootstrap.zip
	@echo "Deploy complete!"

run-local:
	@echo "Starting API locally..."
	@go run ./cmd/api

local-dynamodb:
	@echo "Starting local DynamoDB..."
	@docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
	@echo "DynamoDB running on http://localhost:8000"
	@echo "Create table with: aws dynamodb create-table --cli-input-json file://infrastructure/dynamodb-local.json --endpoint-url http://localhost:8000"

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
