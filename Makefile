SHELL := /bin/bash

clean: ## Clean temp directories
	@echo "Cleaning up temp dirs"
	rm -rf scratch-dir*
	rm -rf ./ab/output/*
	rm -rf dist/

cli-docs: ## Generate auto-generate cli repository documentation
	@echo "Generating cli-docs"
	@go run docs/generate-docs.go

test: ## Run tests
	@echo "running tests"
	@go test -v -cover ./...

test-with-output: ## Run tests with coverage file output
	@echo "running tests with coverage output"
	@go test -v ./... -covermode=count -coverprofile=/tmp/coverage.out
	@go tool cover -func=/tmp/coverage.out -o=/tmp/coverage.out
	@echo "COVERAGE_PERCENT=$$(grep '(statements)' /tmp/coverage.out|awk '{ print $$3 }'|sed 's/\%//g')" > /tmp/percentage

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
