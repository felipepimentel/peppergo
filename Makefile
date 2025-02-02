.PHONY: all build test lint clean validate docs generate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=peppergo
MAIN_PATH=cmd/peppergo/main.go

# Tool commands
GOLINT=golangci-lint
SWAG=swag
PROTOC=protoc

all: validate test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v -race -cover ./...

lint:
	$(GOLINT) run

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

validate:
	./scripts/validate_structure.sh

tidy:
	$(GOMOD) tidy

vendor:
	$(GOMOD) vendor

docs:
	$(SWAG) init -g $(MAIN_PATH)

generate:
	$(GOCMD) generate ./...

# Proto generation
proto:
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

# Development helpers
dev: validate lint test
	@echo "Development checks completed"

# Install development tools
tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Run the application
run:
	$(GOCMD) run $(MAIN_PATH)

# Run tests with coverage report
cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Security scan
security:
	gosec ./...
	$(GOGET) -u github.com/sonatype-nexus-community/nancy
	$(GOLIST) -json -m all | nancy sleuth

# Docker commands
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

# Help command
help:
	@echo "Available commands:"
	@echo "  make all          - Run validate, test, and build"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Clean build files"
	@echo "  make validate     - Validate project structure"
	@echo "  make docs         - Generate API documentation"
	@echo "  make proto        - Generate protobuf code"
	@echo "  make dev          - Run development checks"
	@echo "  make tools        - Install development tools"
	@echo "  make run          - Run the application"
	@echo "  make cover        - Generate test coverage report"
	@echo "  make bench        - Run benchmarks"
	@echo "  make security     - Run security checks"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container" 