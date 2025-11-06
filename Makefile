# Makefile for the chess project

.PHONY: all build test clean

# Default target
all: build

# Build the example CLI application
build:
	@echo "Building chess CLI..."
	@go build -o chess-cli ./examples/main.go
	@echo "Build complete: ./chess-cli"

# Run all unit tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f chess-cli