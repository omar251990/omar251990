#!/bin/bash
################################################################################
# Protei_Bulk - Temporary Files Cleanup Utility
# Cleans up temporary and cache files
################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
TMP_DIR="$BASE_DIR/tmp"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Protei_Bulk Temporary Files Cleanup"
echo "===================================="

# Check if tmp directory exists
if [ ! -d "$TMP_DIR" ]; then
    echo -e "${YELLOW}Warning: Temporary directory not found at $TMP_DIR${NC}"
    exit 0
fi

echo "Cleaning temporary files in $TMP_DIR..."
echo ""

# Count files before cleanup
before_count=$(find "$TMP_DIR" -type f | wc -l)
before_size=$(du -sh "$TMP_DIR" 2>/dev/null | cut -f1)

echo "Before cleanup:"
echo "  Files: $before_count"
echo "  Size: $before_size"
echo ""

# Clean cache directory
if [ -d "$TMP_DIR/cache" ]; then
    echo "Cleaning cache..."
    find "$TMP_DIR/cache" -type f -mtime +7 -delete
fi

# Clean parser directory
if [ -d "$TMP_DIR/parser" ]; then
    echo "Cleaning parser files..."
    find "$TMP_DIR/parser" -type f -mtime +1 -delete
fi

# Clean buffer directory
if [ -d "$TMP_DIR/buffer" ]; then
    echo "Cleaning buffer files..."
    find "$TMP_DIR/buffer" -type f -mtime +1 -delete
fi

# Remove empty directories
find "$TMP_DIR" -type d -empty -delete 2>/dev/null

# Count files after cleanup
after_count=$(find "$TMP_DIR" -type f 2>/dev/null | wc -l)
after_size=$(du -sh "$TMP_DIR" 2>/dev/null | cut -f1)

echo ""
echo "After cleanup:"
echo "  Files: $after_count"
echo "  Size: $after_size"
echo ""

files_removed=$((before_count - after_count))
echo -e "${GREEN}✓ Cleanup complete${NC}"
echo "  Removed: $files_removed files"
