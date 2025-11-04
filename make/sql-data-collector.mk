SQL_ANALYZER_VERSION ?= 0.1.0
SQL_ANALYZER_IMG ?= quay.io/oceanbase/sql-analyzer:${SQL_ANALYZER_VERSION}

.PHONY: sql-analyzer
sql-analyzer: 
	@echo Building sql-analyzer...
	@mkdir -p bin
	@go build -o bin/sql-analyzer cmd/sql-analyzer/main.go

.PHONY: sql-analyzer-image
sql-analyzer-image: 
	$(eval DOCKER_BUILD_ARGS :=)
	$(if $(GOPROXY),$(eval DOCKER_BUILD_ARGS := --build-arg GOPROXY=$(GOPROXY)))
	docker build $(DOCKER_BUILD_ARGS) -t ${SQL_ANALYZER_IMG} -f build/Dockerfile.sql-analyzer .
