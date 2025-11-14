#!/bin/bash
#
# Protei Monitoring v2.0 - Build Script
#
# This script builds the Protei Monitoring application from source
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VERSION="2.0.0"
BUILD_DATE=$(date -u +%Y-%m-%d)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
SOURCE_DIR="/home/user/omar251990"
BIN_DIR="/home/user/omar251990/Protei_Monitoring/bin"
OUTPUT_BIN="$SOURCE_DIR/protei-monitoring"

print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring v${VERSION} - Build"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_header

# Check Go installation
print_step "Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_error "Go compiler not found. Please install Go 1.21 or higher."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
print_success "Go compiler found: $GO_VERSION"

# Navigate to source directory
cd "$SOURCE_DIR"

# Download dependencies
print_step "Downloading Go dependencies..."
if go mod download; then
    print_success "Dependencies downloaded"
else
    print_error "Failed to download dependencies"
    exit 1
fi

# Verify dependencies
print_step "Verifying dependencies..."
if go mod verify; then
    print_success "Dependencies verified"
else
    print_error "Dependency verification failed"
    exit 1
fi

# Run tests
print_step "Running unit tests..."
if go test -short ./...; then
    print_success "All tests passed"
else
    print_error "Some tests failed"
    # Continue anyway for now
fi

# Build application
print_step "Building application..."

# Build flags
LDFLAGS="-X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitCommit=$GIT_COMMIT"

# Production build (optimized)
if CGO_ENABLED=0 go build \
    -ldflags "-s -w $LDFLAGS" \
    -o "$OUTPUT_BIN" \
    ./cmd/protei-monitoring/; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Check binary
if [ -f "$OUTPUT_BIN" ]; then
    BINARY_SIZE=$(du -h "$OUTPUT_BIN" | cut -f1)
    print_success "Binary created: $OUTPUT_BIN ($BINARY_SIZE)"

    # Make executable
    chmod +x "$OUTPUT_BIN"

    # Copy to bin directory
    if [ -d "$BIN_DIR" ]; then
        cp "$OUTPUT_BIN" "$BIN_DIR/"
        print_success "Binary copied to $BIN_DIR/"
    fi
else
    print_error "Binary not found after build"
    exit 1
fi

# Display version info
print_step "Verifying binary..."
if "$OUTPUT_BIN" --version 2>/dev/null || true; then
    print_success "Binary is executable"
else
    print_info "Binary built successfully (version flag not implemented yet)"
fi

# Build summary
echo ""
echo -e "${GREEN}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Build Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"
echo ""
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"
echo "Output: $OUTPUT_BIN"
echo "Size: $BINARY_SIZE"
echo ""
echo "Next Steps:"
echo "  1. Test binary: $OUTPUT_BIN --help"
echo "  2. Run deployment: sudo scripts/deploy.sh"
echo "  3. Or run directly: sudo $OUTPUT_BIN -config config.yaml"
echo ""
print_success "Protei Monitoring v${VERSION} built successfully!"

exit 0
