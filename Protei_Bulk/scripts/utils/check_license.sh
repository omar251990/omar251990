#!/bin/bash
################################################################################
# Protei_Bulk - License Check Utility
# Validates the application license
################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
LICENSE_FILE="$BASE_DIR/config/license.key"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "Protei_Bulk License Check"
echo "=========================="
echo ""

# Check if license file exists
if [ ! -f "$LICENSE_FILE" ]; then
    echo -e "${RED}✗ License file not found${NC}"
    echo "Expected location: $LICENSE_FILE"
    exit 1
fi

echo -e "${BLUE}License File:${NC} $LICENSE_FILE"
echo ""

# Parse license file
while IFS='=' read -r key value; do
    # Skip comments and empty lines
    [[ $key =~ ^#.*$ ]] && continue
    [[ -z $key ]] && continue

    case $key in
        LICENSE_KEY)
            if [ "$value" == "DEMO_LICENSE_KEY_REPLACE_WITH_ACTUAL_KEY" ]; then
                echo -e "${YELLOW}⚠ License:${NC} DEMO LICENSE (Not for production use)"
            else
                echo -e "${GREEN}✓ License:${NC} Valid key present"
            fi
            ;;
        CUSTOMER_ID)
            echo -e "${BLUE}  Customer ID:${NC} $value"
            ;;
        MAX_CONNECTIONS)
            echo -e "${BLUE}  Max Connections:${NC} $value"
            ;;
        MAX_THROUGHPUT)
            echo -e "${BLUE}  Max Throughput:${NC} $value messages/sec"
            ;;
        FEATURES)
            echo -e "${BLUE}  Enabled Features:${NC} $value"
            ;;
        EXPIRY_DATE)
            echo -e "${BLUE}  Expiry Date:${NC} $value"

            # Check if expired
            expiry_seconds=$(date -d "$value" +%s 2>/dev/null)
            current_seconds=$(date +%s)

            if [ -n "$expiry_seconds" ]; then
                if [ $current_seconds -gt $expiry_seconds ]; then
                    echo -e "${RED}  ✗ LICENSE EXPIRED${NC}"
                    exit 1
                else
                    days_remaining=$(( ($expiry_seconds - $current_seconds) / 86400 ))
                    if [ $days_remaining -lt 30 ]; then
                        echo -e "${YELLOW}  ⚠ Expires in $days_remaining days${NC}"
                    else
                        echo -e "${GREEN}  ✓ Valid for $days_remaining days${NC}"
                    fi
                fi
            fi
            ;;
    esac
done < "$LICENSE_FILE"

echo ""
echo -e "${GREEN}✓ License check complete${NC}"
