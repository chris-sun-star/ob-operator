##@ sql-data-collector

COLLECTOR_BINARY_NAME ?= sql-data-collector
COLLECTOR_BUILD_DIR ?= bin/
COLLECTOR_IMG ?= oceanbase/ob-sql-data-collector

.PHONY: build-sql-data-collector
build-sql-data-collector: ## Build sql-data-collector
	@echo "Building $(COLLECTOR_BINARY_NAME)..."
	@go build -o $(COLLECTOR_BUILD_DIR)$(COLLECTOR_BINARY_NAME) ./cmd/sql-data-collector

.PHONY: run-sql-data-collector
run-sql-data-collector: ## Run sql-data-collector in dev mode
	@go run ./cmd/sql-data-collector

.PHONY: docker-build-sql-data-collector
docker-build-sql-data-collector: ## Build docker image for sql-data-collector.
	@echo "Building docker image for $(COLLECTOR_BINARY_NAME)..."
	@docker build -t $(COLLECTOR_IMG):$(VERSION) -f build/Dockerfile.sql-data-collector .
