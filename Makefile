.PHONY: all build test bench lint fmt vet cover install clean fetch-data help

BIN       := naics
CMD       := ./cmd/naics
COVERFILE := coverage.out

all: lint test build ## Run lint, test, and build

build: ## Build the CLI binary
	go build -o $(BIN) $(CMD)

install: ## Install the CLI to $GOPATH/bin
	go install $(CMD)

test: ## Run all tests with race detector
	go test -race -count=1 ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./internal/core/

cover: ## Run tests with coverage report
	go test -race -coverprofile=$(COVERFILE) ./...
	go tool cover -func=$(COVERFILE)
	@echo ""
	@echo "HTML report: go tool cover -html=$(COVERFILE)"

lint: vet fmt ## Run all linters (vet + fmt check)

vet: ## Run go vet
	go vet ./...

fmt: ## Check formatting (fails if files need formatting)
	@test -z "$$(gofmt -l .)" || { echo "Files need formatting:"; gofmt -l .; exit 1; }

fetch-data: ## Refresh embedded NAICS data from upstream
	go run ./internal/cmd/fetchdata

clean: ## Remove build artifacts
	rm -f $(BIN) $(COVERFILE)
	go clean -cache -testcache

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'
