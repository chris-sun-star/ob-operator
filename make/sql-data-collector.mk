.PHONY: build-sql-data-collector
build-sql-data-collector: ## Build sql-data-collector binary.
	@echo Building sql-data-collector...
	@mkdir -p bin
	@go build -o bin/sql-data-collector cmd/sql-data-collector/main.go

.PHONY: docker-build-sql-data-collector
docker-build-sql-data-collector: ## Build docker image for sql-data-collector.
	@echo Building docker image for sql-data-collector...
	@docker build -t sql-data-collector:latest -f build/Dockerfile.sql-data-collector . --build-arg GOPROXY=$(GOPROXY)
