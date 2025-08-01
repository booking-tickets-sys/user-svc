# Makefile for user-svc

.PHONY: all build test clean

# Default target
all: build

# Build the application
build: proto-gen
	@echo "Building user-svc..."
	go build -o bin/user-svc ./main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run the application
run: build
	@echo "Running user-svc..."
	./bin/user-svc

# Development setup
dev-setup: deps
	@echo "Development setup completed!"

# Generate mocks
mock:
	@echo "Generating mocks..."
	mockgen -source=internal/app/service/user.go -destination=mocks/mock_interfaces.go -package=mocks

# Clean mocks
mock-clean:
	@echo "Cleaning mocks..."
	rm -rf mocks/*

# Update proto submodule
proto-update:
	@echo "Updating proto submodule..."
	git submodule update --remote submodules/proto
	@echo "Proto submodule updated!"

# Generate protobuf files from submodule
proto-gen:
	@echo "Generating protobuf files from submodule..."
	@mkdir -p pb
	cd submodules/proto && make gen-grpc
	cp submodules/proto/pb/*.go pb/
	@echo "Protobuf files generated!"

# Clean protobuf files
proto-clean:
	@echo "Cleaning protobuf files..."
	rm -rf pb/*.go
	cd submodules/proto && make clean
	@echo "Protobuf files cleaned!"

# Setup proto (update submodule and generate files)
proto-setup: proto-update proto-gen
	@echo "Proto setup completed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Build the application (default)"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  run          - Build and run the application"
	@echo "  dev-setup    - Setup development environment"
	@echo "  mock         - Generate mocks for testing"
	@echo "  mock-clean   - Clean generated mocks"
	@echo "  proto-update - Update proto submodule"
	@echo "  proto-gen    - Generate protobuf files from submodule"
	@echo "  proto-clean  - Clean protobuf files"
	@echo "  proto-setup  - Update submodule and generate files"
	@echo "  help         - Show this help message"