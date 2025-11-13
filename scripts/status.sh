#!/bin/bash
# Protei_Monitoring Status Script

INSTALL_DIR="/usr/protei/Protei_Monitoring"
PID_FILE="$INSTALL_DIR/protei-monitoring.pid"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Protei_Monitoring Status         ${NC}"
echo -e "${GREEN}====================================${NC}"
echo

# Check if running
if [ ! -f "$PID_FILE" ]; then
    echo -e "${RED}Status: NOT RUNNING${NC}"
    echo "PID file not found: $PID_FILE"
    exit 1
fi

PID=$(cat "$PID_FILE")

if ! ps -p "$PID" > /dev/null 2>&1; then
    echo -e "${RED}Status: NOT RUNNING${NC}"
    echo "PID file exists but process $PID is not running"
    exit 1
fi

# Get process info
echo -e "${GREEN}Status: RUNNING${NC}"
echo "PID: $PID"
echo

# Get uptime
START_TIME=$(ps -p "$PID" -o lstart=)
echo "Started: $START_TIME"

# Get resource usage
echo
echo "Resource Usage:"
ps -p "$PID" -o pid,ppid,%cpu,%mem,vsz,rss,tty,stat,start,time,cmd

# Check HTTP health endpoint
echo
echo "Health Check:"
if command -v curl &> /dev/null; then
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
    if [ "$HTTP_STATUS" == "200" ]; then
        echo -e "${GREEN}✓ Health endpoint responding (HTTP $HTTP_STATUS)${NC}"
        echo
        echo "API Response:"
        curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health
    else
        echo -e "${YELLOW}⚠ Health endpoint returned HTTP $HTTP_STATUS${NC}"
    fi
else
    echo "curl not available, skipping health check"
fi

echo
echo "Dashboard: http://localhost:8080"
