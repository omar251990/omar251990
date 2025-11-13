#!/bin/bash
# Protei_Monitoring Stop Script

set -e

INSTALL_DIR="/usr/protei/Protei_Monitoring"
PID_FILE="$INSTALL_DIR/protei-monitoring.pid"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}====================================${NC}"
echo -e "${YELLOW}  Protei_Monitoring Stop Script    ${NC}"
echo -e "${YELLOW}====================================${NC}"
echo

# Check if running
if [ ! -f "$PID_FILE" ]; then
    echo -e "${YELLOW}Protei_Monitoring is not running (PID file not found)${NC}"
    exit 0
fi

PID=$(cat "$PID_FILE")

if ! ps -p "$PID" > /dev/null 2>&1; then
    echo -e "${YELLOW}Protei_Monitoring is not running (process not found)${NC}"
    rm -f "$PID_FILE"
    exit 0
fi

# Stop the application
echo "Stopping Protei_Monitoring (PID: $PID)..."

# Send SIGTERM for graceful shutdown
kill -TERM "$PID"

# Wait for process to stop (max 30 seconds)
TIMEOUT=30
ELAPSED=0

while ps -p "$PID" > /dev/null 2>&1; do
    if [ $ELAPSED -ge $TIMEOUT ]; then
        echo -e "${RED}Process did not stop gracefully, forcing shutdown...${NC}"
        kill -KILL "$PID"
        sleep 1
        break
    fi

    echo -n "."
    sleep 1
    ELAPSED=$((ELAPSED + 1))
done

echo

# Remove PID file
rm -f "$PID_FILE"

echo -e "${GREEN}✓ Protei_Monitoring stopped successfully${NC}"
