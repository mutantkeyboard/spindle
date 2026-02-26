.PHONY: test lint vet fmt check docker-test docker-build clean

## Testing

test: ## Run tests with race detector
	go test -race -v ./...

test-cover: ## Run tests with coverage
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

## Code quality

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code
	gofmt -w .

vet: ## Run go vet
	go vet ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

## Docker

docker-test: ## Run tests in Docker
	docker build -f Dockerfile.test -t spindle-test .
	docker run --rm spindle-test

docker-build: ## Build dev Docker image
	docker build -t spindle-dev .

## Cleanup

clean: ## Remove build artifacts
	rm -f coverage.out
	rm -rf dist/

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
