#!/bin/bash
# Protei_Monitoring Start Script

set -e

INSTALL_DIR="/usr/protei/Protei_Monitoring"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/configs"
LOG_DIR="$INSTALL_DIR/logs"
PID_FILE="$INSTALL_DIR/protei-monitoring.pid"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Protei_Monitoring Start Script   ${NC}"
echo -e "${GREEN}====================================${NC}"
echo

# Check if already running
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}Protei_Monitoring is already running (PID: $PID)${NC}"
        exit 1
    else
        echo -e "${YELLOW}Removing stale PID file${NC}"
        rm -f "$PID_FILE"
    fi
fi

# Create required directories
mkdir -p "$LOG_DIR" "$INSTALL_DIR/out/events" "$INSTALL_DIR/out/cdr" "$INSTALL_DIR/out/diagrams" "$INSTALL_DIR/ingest"

# Start the application
echo "Starting Protei_Monitoring..."
cd "$INSTALL_DIR"

nohup "$BIN_DIR/protei-monitoring" -config="$CONFIG_DIR/config.yaml" > "$LOG_DIR/console.log" 2>&1 &
PID=$!

# Save PID
echo $PID > "$PID_FILE"

# Wait a moment and check if it's running
sleep 2

if ps -p "$PID" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Protei_Monitoring started successfully (PID: $PID)${NC}"
    echo "  Log file: $LOG_DIR/console.log"
    echo "  Dashboard: http://localhost:8080"
else
    echo -e "${RED}✗ Failed to start Protei_Monitoring${NC}"
    echo "  Check logs at: $LOG_DIR/console.log"
    rm -f "$PID_FILE"
    exit 1
fi

echo
echo -e "${GREEN}Application is ready!${NC}"
