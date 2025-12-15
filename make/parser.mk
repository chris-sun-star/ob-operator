##@ Parser Generation

ANTLR_BUILDER_IMAGE = oceanbase/antlr:latest

.PHONY: build-antlr-image
build-antlr-image: ## Build the docker image used for parser generation
	docker build --network=host -t $(ANTLR_BUILDER_IMAGE) -f build/Dockerfile.antlr build/

.PHONY: generate-parser
generate-parser: build-antlr-image ## Generate Go parser code from ANTLR4 grammar files using Docker
	@echo "Generating Parser..."
	@mkdir -p internal/sql-analyzer/parser/mysql
	@docker run --rm -u $$(id -u):$$(id -g) -v $$(pwd):/work $(ANTLR_BUILDER_IMAGE) -Dlanguage=Go -o /work/internal/sql-analyzer/parser/mysql -visitor -package mysql /work/obparser/obmysql/sql/OBLexer.g4 /work/obparser/obmysql/sql/OBParser.g4
	@echo "Parser generation complete."
