.PHONY: help build test clean install lint fmt vet coverage release-dry

# Variables
BINARY_NAME=ygg
GO_FILES=$(shell find . -name '*.go' -type f)
MAIN_PACKAGE=./cmd/ygg
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Default target
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	go build ${LDFLAGS} -o ${BINARY_NAME} ${MAIN_PACKAGE}

test: ## Run tests
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	rm -f ${BINARY_NAME}
	rm -f coverage.out coverage.html
	rm -rf dist/

install: ## Install the binary to GOPATH/bin
	go install ${LDFLAGS} ${MAIN_PACKAGE}

lint: ## Run linters
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w ${GO_FILES}

vet: ## Run go vet
	go vet ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

update-deps: ## Update dependencies
	go get -u ./...
	go mod tidy

# Development helpers
run: build ## Build and run
	./${BINARY_NAME}

watch: ## Watch for changes and rebuild
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Release targets
release-dry: ## Dry run of goreleaser
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --snapshot --clean --skip-publish

release-snapshot: ## Create a snapshot release
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --snapshot --clean

# Completion targets
completions: build ## Generate shell completions
	mkdir -p completions
	./${BINARY_NAME} completion bash > completions/ygg.bash
	./${BINARY_NAME} completion zsh > completions/ygg.zsh
	./${BINARY_NAME} completion fish > completions/ygg.fish

install-completions: completions ## Install shell completions (requires sudo on some systems)
	@echo "Installing completions..."
	@echo "For bash: source completions/ygg.bash"
	@echo "For zsh: cp completions/ygg.zsh ~/.zsh/completions/_ygg"
	@echo "For fish: cp completions/ygg.fish ~/.config/fish/completions/"

# Quick commands for development
.PHONY: dev
dev: fmt vet test build ## Run fmt, vet, test, and build

.PHONY: ci
ci: deps fmt vet lint test ## Run all CI checks

.PHONY: all
all: clean deps fmt vet lint test build ## Run everything