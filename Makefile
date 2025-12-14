SHELL := /bin/bash
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: help install-tools lint lint-go lint-md lint-yaml test test-coverage clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install all linting tools
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.62.2
	@echo "Installing markdownlint-cli..."
	@npm install -g markdownlint-cli
	@echo "Installing yamllint..."
	@pip3 install yamllint
	@echo "All tools installed successfully!"

lint: lint-go lint-md lint-yaml ## Run all linters

lint-go: ## Run Go linter (golangci-lint)
	@echo "Running golangci-lint..."
	@cd backend && golangci-lint run --config ../.golangci.yml

lint-md: ## Run Markdown linter
	@echo "Running markdownlint..."
	@markdownlint '**/*.md' --config .markdownlint.json --ignore node_modules --ignore frontend/node_modules

lint-yaml: ## Run YAML linter
	@echo "Running yamllint..."
	@yamllint -c .yamllint.yml .

test: ## Run all Go tests
	@echo "Running Go tests..."
	@cd backend && go test -v ./...

test-coverage: ## Run Go tests with coverage
	@echo "Running Go tests with coverage..."
	@cd backend && go test -v -race -coverprofile=coverage.out ./...
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: backend/coverage.html"

test-short: ## Run Go tests without verbose output
	@cd backend && go test ./...

build-backend: ## Build the backend
	@echo "Building backend..."
	@cd backend && go build -v -o server ./cmd/server

build-frontend: ## Build the frontend
	@echo "Building frontend..."
	@cd frontend && npm run build

build: build-backend build-frontend ## Build both backend and frontend

run-backend: ## Run the backend server
	@cd backend && go run ./cmd/server -config config.yml

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@cd backend && go fmt ./...
	@cd backend && goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@cd backend && go vet ./...

tidy: ## Tidy Go modules
	@cd backend && go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f backend/server
	@rm -f backend/coverage.out
	@rm -f backend/coverage.html
	@rm -rf backend/dist
	@rm -rf frontend/dist

ci: lint test ## Run CI checks locally (lint + test)

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
