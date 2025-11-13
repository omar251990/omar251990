#!/bin/bash
# Protei_Monitoring Restart Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Restarting Protei_Monitoring..."
echo

# Stop
"$SCRIPT_DIR/stop.sh"

# Wait a moment
sleep 2

# Start
"$SCRIPT_DIR/start.sh"
