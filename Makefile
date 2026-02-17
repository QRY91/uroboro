# Makefile for uroboro - Trinity Development Assistant
# Comprehensive testing and build automation

.PHONY: help build test test-unit test-integration test-all clean lint fmt check deps verify ci install

# Default target
help: ## Show this help message
	@echo "uroboro - Trinity Development Assistant"
	@echo "======================================"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the uroboro binary
	@echo "🔨 Building uroboro..."
	go build -o uroboro ./cmd/uroboro
	@echo "✅ Build complete: ./uroboro"

build-release: ## Build optimized release binary
	@echo "🔨 Building release binary..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o uroboro ./cmd/uroboro
	@echo "✅ Release build complete: ./uroboro"

# Test targets
test: test-unit ## Run unit tests (default test target)

test-unit: ## Run all unit tests
	@echo "🧪 Running unit tests..."
	go test -v ./...
	@echo "✅ Unit tests complete"

test-integration: build ## Run integration tests (requires binary)
	@echo "🧪 Running integration tests..."
	./scripts/test-integration.sh
	@echo "✅ Integration tests complete"

test-all: test-unit test-integration ## Run all tests (unit + integration)
	@echo "🎉 All tests complete!"

# Code quality targets
lint: ## Run code linting
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, running basic checks..."; \
		go vet ./...; \
		gofmt -l . | grep -E '.*\.go$$' && exit 1 || true; \
	fi
	@echo "✅ Linting complete"

fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Formatting complete"

check: ## Run code quality checks
	@echo "🔍 Running quality checks..."
	go vet ./...
	gofmt -l . | grep -E '.*\.go$$' && echo "❌ Code not formatted" && exit 1 || true
	@echo "✅ Quality checks complete"

deps: ## Verify and clean dependencies
	@echo "📦 Checking dependencies..."
	go mod verify
	go mod tidy
	@echo "✅ Dependencies verified and cleaned"

# Verification targets
verify: check deps ## Run all verification steps
	@echo "✅ All verifications complete"

# CI target (comprehensive)
ci: verify test-all ## Run complete CI pipeline
	@echo "🎉 CI pipeline complete - ready for commit!"

# Development targets
dev-setup: ## Setup development environment
	@echo "🛠️  Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📥 Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "✅ Development environment ready"

install: build ## Install uroboro binary to $GOPATH/bin
	@echo "📦 Installing uroboro..."
	cp uroboro $$(go env GOPATH)/bin/
	@echo "✅ uroboro installed to $$(go env GOPATH)/bin/"

# Cleanup targets
clean: ## Clean build artifacts and test files
	@echo "🧹 Cleaning up..."
	rm -f uroboro
	rm -f /tmp/uroboro_test_*.sqlite
	rm -rf /tmp/test_integration_project
	@echo "✅ Cleanup complete"

clean-all: clean ## Clean everything including dependencies
	@echo "🧹 Deep cleaning..."
	go clean -cache -testcache -modcache
	@echo "✅ Deep cleanup complete"

# Database targets
db-backup: ## Backup local uroboro database
	@echo "💾 Backing up database..."
	@if [ -f ~/.local/share/uroboro/uroboro.sqlite ]; then \
		cp ~/.local/share/uroboro/uroboro.sqlite ./backups/uroboro-backup-$$(date +%Y-%m-%d-%H%M%S).sqlite; \
		echo "✅ Database backed up to ./backups/"; \
	else \
		echo "⚠️  No database found at ~/.local/share/uroboro/uroboro.sqlite"; \
	fi

# Development helpers
demo: build ## Run a quick demo of uroboro functionality  
	@echo "🎬 Running uroboro demo..."
	./uroboro capture "Demo: Testing uroboro functionality from Makefile"
	./uroboro status --days 1
	@echo "✅ Demo complete"

# Quick development cycle
quick: fmt test-unit build ## Quick development cycle: format, test, build
	@echo "⚡ Quick cycle complete - ready for testing!"

# Pre-commit hook simulation
pre-commit: verify test-unit ## Simulate pre-commit checks
	@echo "🔐 Pre-commit checks complete - ready to commit!"

# Documentation targets
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	go doc -all ./... > docs/api.md 2>/dev/null || echo "⚠️  go doc failed, skipping"
	@echo "✅ Documentation updated"

# Coverage targets
coverage: ## Run tests with coverage
	@echo "📊 Running coverage analysis..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Performance targets
bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "✅ Benchmarks complete"

# Git helpers
git-hooks: ## Install git hooks for automated testing
	@echo "🪝 Installing git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/bash\nmake pre-commit' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"

# Style distillation
style-distill: build ## Extract style signals from git + uroboro across repos
	@echo "Extracting style signals..."
	./scripts/distill-multi.sh --days 180 --correlate \
		--out ~/.local/share/uroboro/style-data/ \
		$(STYLE_REPOS)
	@echo "Run a Claude Code session with scripts/style-analysis-prompt.md to analyze"

# Default target when no target specified
.DEFAULT_GOAL := help