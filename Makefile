.PHONY: all build test clean lint examples

all: build test

build:
	@echo "Building go-mcp library..."
	go build ./pkg/...
	go build ./examples/...

test:
	@echo "Running tests..."
	go test -v ./pkg/...

clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f examples/basic/basic
	rm -f examples/openai/openai

deps:
	@echo "Installing dependencies..."
	go get github.com/google/uuid
	go get golang.org/x/lint/golint

help:
	@echo "Available targets:"
	@echo "  all         - Build and test the library (default)"
	@echo "  build       - Build the library and examples"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  help        - Show this help information" 