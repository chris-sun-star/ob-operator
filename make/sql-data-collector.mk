.PHONY: build-sql-data-collector
build-sql-data-collector: ## Build sql-data-collector binary.
	@echo Building sql-data-collector...
	@go build -o $(BIN_DIR)/sql-data-collector $(ROOT_DIR)/cmd/sql-data-collector

.PHONY: docker-build-sql-data-collector
docker-build-sql-data-collector: ## Build docker image for sql-data-collector.
	@echo Building docker image for sql-data-collector...
	@docker build -t sql-data-collector:latest -f build/Dockerfile.sql-data-collector . --build-arg GOPROXY=$(GOPROXY)