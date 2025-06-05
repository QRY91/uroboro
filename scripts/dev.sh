#!/bin/bash
set -e

echo "🚀 Uroboro Local Development Pipeline"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 1. Unit Tests
print_status "Running unit tests..."
if go test -v ./internal/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# 2. Coverage Tests
print_status "Running tests with coverage..."
if go test -race -coverprofile=coverage.out -covermode=atomic ./internal/...; then
    print_success "Coverage tests passed"
    go tool cover -html=coverage.out -o coverage.html
    echo "📊 Coverage report: coverage.html"
else
    print_error "Coverage tests failed"
    exit 1
fi

# 3. Build Both Binaries
print_status "Building binaries..."
if go build -o uroboro ./cmd/uroboro && go build -o uro ./cmd/uroboro; then
    print_success "Binaries built successfully"
else
    print_error "Build failed"
    exit 1
fi

# 4. Integration Tests
print_status "Running integration tests..."

# Test basic functionality
./uroboro capture "Local dev test capture - ensuring binary works" || exit 1
./uro -s || exit 1
./uro -c "Short flag test from local dev" || exit 1

# Test full workflow
./uroboro capture "Local integration test capture" || exit 1
./uro -c "Short capture test" || exit 1

# Test status commands
./uroboro status || exit 1
./uro -s || exit 1

# Verify XDG compliance
if [[ -d ~/.local/share/uroboro/daily ]] && [[ -f ~/.local/share/uroboro/daily/$(date +%Y-%m-%d).md ]]; then
    print_success "XDG compliance verified"
else
    print_error "XDG compliance check failed"
    exit 1
fi

# 5. Quality Gate
print_status "Running quality gate checks..."

# Ensure all 3 commands exist and work
./uroboro capture "Quality gate test" || exit 1
./uroboro status || exit 1
./uroboro publish --help || exit 1

print_success "Quality gate passed - no regressions detected!"

# 6. Landing Page Check (bonus)
print_status "Checking landing page..."
if [[ -f landing-page/index.html ]] && [[ -f landing-page/style.css ]]; then
    print_success "Landing page files present"
    echo "🌐 Run 'npm run dev' in landing-page/ to preview"
else
    echo "⚠️  Landing page files missing"
fi

echo ""
echo "🎉 All local checks passed! Ready to push."
echo "💡 To preview landing page: cd landing-page && npm run dev"
echo "📊 Coverage report: open coverage.html" 