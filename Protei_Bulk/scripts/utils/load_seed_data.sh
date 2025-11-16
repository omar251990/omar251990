#!/bin/bash
################################################################################
# Protei_Bulk - Load Seed Data Script
# Loads initial demo data into the database
################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Database credentials
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-protei}"
DB_PASSWORD="${DB_PASSWORD:-elephant}"
DB_NAME="${DB_NAME:-protei_bulk}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Protei_Bulk - Loading Seed Data"
echo "================================"
echo ""

# Check if seed data file exists
SEED_FILE="$BASE_DIR/database/seed_data.sql"

if [ ! -f "$SEED_FILE" ]; then
    echo -e "${RED}Error: Seed data file not found: $SEED_FILE${NC}"
    exit 1
fi

echo "Database: $DB_NAME@$DB_HOST:$DB_PORT"
echo "Seed File: $SEED_FILE"
echo ""

read -p "This will add demo data to the database. Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Load seed data
export PGPASSWORD="$DB_PASSWORD"

echo "Loading seed data..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SEED_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Seed data loaded successfully${NC}"
    echo ""
    echo "Default credentials:"
    echo "  Username: admin"
    echo "  Password: Admin@123"
    echo "  (Change on first login)"
    echo ""
else
    echo -e "${RED}✗ Failed to load seed data${NC}"
    exit 1
fi

unset PGPASSWORD
