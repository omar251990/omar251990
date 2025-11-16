#!/bin/bash
################################################################################
# Protei_Bulk - Log Rotation Utility
# Rotates and compresses old log files
################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
LOG_DIR="$BASE_DIR/logs"
ARCHIVE_DIR="$LOG_DIR/archive"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Configuration
MAX_SIZE_MB=100
RETENTION_DAYS=90

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Protei_Bulk Log Rotation"
echo "========================="

# Create archive directory
mkdir -p "$ARCHIVE_DIR"

# Rotate logs
cd "$LOG_DIR" || exit 1

for logfile in *.log; do
    [ -f "$logfile" ] || continue

    # Check file size
    size=$(du -m "$logfile" | cut -f1)

    if [ "$size" -gt "$MAX_SIZE_MB" ]; then
        echo "Rotating $logfile (${size}MB)..."

        # Copy and compress
        cp "$logfile" "$ARCHIVE_DIR/${logfile%.log}_${TIMESTAMP}.log"
        gzip "$ARCHIVE_DIR/${logfile%.log}_${TIMESTAMP}.log"

        # Truncate original
        > "$logfile"

        echo -e "${GREEN}✓ Rotated $logfile${NC}"
    fi
done

# Clean old archives
echo ""
echo "Cleaning archives older than $RETENTION_DAYS days..."
find "$ARCHIVE_DIR" -name "*.log.gz" -mtime +$RETENTION_DAYS -delete
echo -e "${GREEN}✓ Cleanup complete${NC}"

echo ""
echo "Current log sizes:"
du -h "$LOG_DIR"/*.log 2>/dev/null | sed 's/^/  /'
