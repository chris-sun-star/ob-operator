##@ sql-data-collector

COLLECTOR_BINARY_NAME ?= sql-data-collector
COLLECTOR_BUILD_DIR ?= bin/

.PHONY: build-sql-data-collector
build-sql-data-collector: ## Build sql-data-collector
	@echo "Building $(COLLECTOR_BINARY_NAME)..."
	@go build -o $(COLLECTOR_BUILD_DIR)$(COLLECTOR_BINARY_NAME) ./cmd/sql-data-collector

.PHONY: run-sql-data-collector
run-sql-data-collector: ## Run sql-data-collector in dev mode
	@go run ./cmd/sql-data-collector
