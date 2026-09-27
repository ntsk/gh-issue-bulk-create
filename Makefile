.PHONY: build test lint clean run dry-run

# Default target
all: lint test build

# Build target
build:
	@echo "Building gh-issue-bulk-create..."
	@go build -o gh-issue-bulk-create

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linters
lint:
	@echo "Running staticcheck..."
	@go tool staticcheck ./...
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running go fmt..."
	@go fmt ./...

# Cleanup
clean:
	@echo "Cleaning up..."
	@rm -f gh-issue-bulk-create

# Example run (dry-run mode)
dry-run:
	@echo "Running with dry-run mode..."
	@go run . --template sample-template.md --csv sample-data.csv --dry-run

# Install as GitHub CLI extension
install:
	@echo "Installing as GitHub CLI extension..."
	@go build -o gh-issue-bulk-create
	@gh extension install .

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Run lint, test and build (default)"
	@echo "  build      - Build the binary"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  clean      - Remove built binary"
	@echo "  dry-run    - Run with dry-run mode"
	@echo "  install    - Install as GitHub CLI extension"
	@echo "  help       - Show this help message"
