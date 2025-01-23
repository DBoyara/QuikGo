PROGRAM_NAME = QuikGo

.PHONY: help clean dep test build lint go-lint

.DEFAULT_GOAL := help

help: ## Display this help screen.
	@echo "Makefile available targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  * \033[36m%-15s\033[0m %s\n", $$1, $$2}'

clean: ## Clean build directory.
	rm -f ./bin/${PROGRAM_NAME}
	rmdir ./bin

dep: ## Download the dependencies.
	go mod download

test: dep ## Run tests
	go test -v -cover ./...

go-lint: ## Run linter
	@docker run --rm -it -v `pwd`:/go/src/QuikGo -w /go/src/QuikGo golangci/golangci-lint golangci-lint run

lint: ## Run linter
	go vet ./...
