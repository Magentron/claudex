.PHONY: all build clean install help test

# Default target
all: build

# Build the executable
build:
	@echo "Building claudex..."
	@go build -o claudex main.go
	@echo "✓ Built: claudex"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "✓ Dependencies installed"

# Install the binary
install: build
	@echo "Installing claudex..."
	@chmod +x claudex
	@echo "✓ Installed: claudex"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f claudex
	@echo "✓ Cleaned"

# Run the program
run: build
	@./claudex

# Show help
help:
	@echo "Available targets:"
	@echo "  make all      - Build the executable (default)"
	@echo "  make build    - Build claudex"
	@echo "  make deps     - Install/update dependencies"
	@echo "  make install  - Build and mark as executable"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make run      - Build and run"
	@echo "  make help     - Show this help"
