SQL_COLLECTOR_VERSION ?= 0.1.0
SQL_COLLECTOR_IMG ?= quay.io/oceanbase/sql-data-collector:${SQL_COLLECTOR_VERSION}

.PHONY: build-sql-data-collector
build-sql-data-collector: ## Build sql-data-collector binary.
	@echo Building sql-data-collector...
	@mkdir -p bin
	@CGO_ENABLED=0 go build -o bin/sql-data-collector cmd/sql-data-collector/main.go

.PHONY: docker-build-sql-data-collector
docker-build-sql-data-collector: ## build sql-data-collector image
	$(eval DOCKER_BUILD_ARGS :=)
	$(if $(GOPROXY),$(eval DOCKER_BUILD_ARGS := --build-arg GOPROXY=$(GOPROXY)))
	docker build $(DOCKER_BUILD_ARGS) -t ${SQL_COLLECTOR_IMG} -f build/Dockerfile.sql-data-collector .
