# Makefile for go-project-generator (the `gen` CLI).

BINARY := gen
BIN_DIR := bin

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the gen binary into bin/
	@echo "Building $(BINARY)..."
	go build -o $(BIN_DIR)/$(BINARY) .

.PHONY: install
install: ## Install gen into $(go env GOPATH)/bin
	go install .

.PHONY: run
run: ## Run gen (pass args via ARGS='new myapp --module github.com/me/myapp')
	go run . $(ARGS)

.PHONY: test
test: ## Run all tests including the network e2e (-race)
	go test -race -count=1 -timeout 5m ./...

.PHONY: test-short
test-short: ## Run unit tests only (skips the network e2e)
	go test -short -race -count=1 ./...

.PHONY: e2e
e2e: ## Run the end-to-end test (fetches the template from GitHub)
	go test -run TestEndToEnd -count=1 -timeout 5m -v .

.PHONY: cover
cover: ## Run unit tests with a coverage summary
	go test -short -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (installs it if missing)
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out

.PHONY: ci
ci: vet lint test ## Run the CI checks locally

.PHONY: all
all: fmt vet lint test build ## Format, vet, lint, test, and build
