.PHONY: all build clean install install-project uninstall help test run deps

# Configuration
CONFIG_DIR = $(HOME)/.config/claudex
BIN_DIR = $(HOME)/.local/bin

# Default target
all: build

# Build the executable
build:
	@echo "Building claudex..."
	@go build -o claudex
	@echo "✓ Built: claudex"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "✓ Dependencies installed"

# Install claudex for current user
install: build
	@echo "Installing claudex..."
	@mkdir -p $(CONFIG_DIR) $(BIN_DIR)
	@cp -r profiles $(CONFIG_DIR)/
	@cp -r .claude/hooks $(CONFIG_DIR)/
	@ln -sf $(CURDIR)/claudex $(BIN_DIR)/claudex
	@echo "✓ Installed to $(CONFIG_DIR)"
	@echo "✓ Binary linked at $(BIN_DIR)/claudex"
	@if ! echo "$$PATH" | grep -q "$(BIN_DIR)"; then \
		echo "⚠ Add to your shell config: export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
	fi

# Uninstall claudex
uninstall:
	@echo "Uninstalling claudex..."
	@rm -rf $(CONFIG_DIR)
	@rm -f $(BIN_DIR)/claudex
	@echo "✓ Uninstalled"

# Install to current project directory
install-project:
	@echo "Installing claudex to current project..."
	@mkdir -p .claude
	@if [ -d "$(CONFIG_DIR)/profiles" ]; then \
		cp -r $(CONFIG_DIR)/profiles .claude/; \
		echo "✓ Copied profiles from $(CONFIG_DIR)"; \
	elif [ -d "profiles" ]; then \
		cp -r profiles .claude/; \
		echo "✓ Copied profiles from local directory"; \
	else \
		echo "✗ No profiles directory found"; \
		exit 1; \
	fi
	@if [ -d "$(CONFIG_DIR)/hooks" ]; then \
		cp -r $(CONFIG_DIR)/hooks .claude/; \
		echo "✓ Copied hooks from $(CONFIG_DIR)"; \
	elif [ -d ".claude/hooks" ]; then \
		echo "✓ Hooks already exist in .claude/"; \
	else \
		echo "⚠ No hooks directory found"; \
	fi
	@echo "✓ Project installation complete"

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
	@echo "  make all             - Build the executable (default)"
	@echo "  make build           - Build claudex"
	@echo "  make deps            - Install/update dependencies"
	@echo "  make install         - Install claudex to ~/.local/bin and ~/.config/claudex"
	@echo "  make uninstall       - Remove claudex installation"
	@echo "  make install-project - Install profiles/hooks to current project .claude/"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make run             - Build and run"
	@echo "  make help            - Show this help"
