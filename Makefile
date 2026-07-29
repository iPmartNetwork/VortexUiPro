.PHONY: build run test lint clean dev help

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
GOMOD=$(GOCMD) mod
BINARY_NAME=vortexuipro
BUILD_DIR=./build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the panel binary
	cd cmd/panel && $(GOBUILD) -o ../../$(BUILD_DIR)/$(BINARY_NAME) .

run: ## Run the panel in development mode
	cd cmd/panel && $(GORUN) . --dev

test: ## Run all tests
	$(GOTEST) -v -race -count=1 ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)/

tidy: ## Tidy Go modules
	$(GOMOD) tidy

dev: ## Run in development mode with hot reload (requires air)
	air --cmd "go run ./cmd/panel --dev"

frontend-dev: ## Run frontend dev server
	cd web && npm run dev

frontend-build: ## Build frontend
	cd web && npm run build

docker-build: ## Build Docker image
	docker build -t vortexuipro:latest -f deploy/Dockerfile .

db-init: ## Initialize database
	cd cmd/panel && $(GORUN) . --init-db

db-migrate: ## Run database migrations
	cd cmd/panel && $(GORUN) . --migrate
