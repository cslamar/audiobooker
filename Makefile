clean: ## Clean temp directories
	@echo "Cleaning up temp dirs"
	rm -rf scratch-dir*
	rm -rf ./ab/output/*

cli-docs: ## Generate auto-generate cli repository documentation
	@echo "Generating cli-docs"
	@go run docs/generate-docs.go

test: ## Run tests
	@echo "running tests"
	@go test -v -cover ./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
