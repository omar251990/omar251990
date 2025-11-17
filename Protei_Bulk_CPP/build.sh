#!/bin/bash
#
# Build script for Protei_Bulk C++
#

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Protei_Bulk C++ Build Script                   ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo

# Parse arguments
BUILD_TYPE="Release"
CLEAN=false
INSTALL=false
RUN_TESTS=false
JOBS=$(nproc)

while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            BUILD_TYPE="Debug"
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --install)
            INSTALL=true
            shift
            ;;
        --test)
            RUN_TESTS=true
            shift
            ;;
        --jobs)
            JOBS="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--debug] [--clean] [--install] [--test] [--jobs N]"
            exit 1
            ;;
    esac
done

# Clean build directory
if [ "$CLEAN" = true ]; then
    echo -e "${YELLOW}[1/5] Cleaning build directory...${NC}"
    rm -rf build
fi

# Create build directory
echo -e "${YELLOW}[2/5] Creating build directory...${NC}"
mkdir -p build
cd build

# Configure with CMake
echo -e "${YELLOW}[3/5] Configuring with CMake (Build Type: $BUILD_TYPE)...${NC}"
cmake .. -DCMAKE_BUILD_TYPE=$BUILD_TYPE

# Build
echo -e "${YELLOW}[4/5] Building with $JOBS parallel jobs...${NC}"
make -j${JOBS}

# Run tests
if [ "$RUN_TESTS" = true ]; then
    echo -e "${YELLOW}[5/5] Running tests...${NC}"
    ctest --output-on-failure
fi

# Install
if [ "$INSTALL" = true ]; then
    echo -e "${YELLOW}[5/5] Installing...${NC}"
    sudo make install
fi

echo
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Build completed successfully!                   ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo
echo -e "Binary location: ${GREEN}$(pwd)/bin/protei_bulk${NC}"
echo
echo "To run the application:"
echo "  cd build && ./bin/protei_bulk"
echo
echo "To run with custom config:"
echo "  cd build && ./bin/protei_bulk /path/to/config/app.conf"
echo
