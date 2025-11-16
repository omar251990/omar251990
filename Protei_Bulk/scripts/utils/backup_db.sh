#!/bin/bash
################################################################################
# Protei_Bulk - Database Backup Utility
# Creates a backup of the database
################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
CONFIG_DIR="$BASE_DIR/config"
BACKUP_DIR="/var/backups/protei_bulk"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Protei_Bulk Database Backup"
echo "============================"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Read database configuration
# TODO: Parse config/db.conf for database credentials
DB_TYPE="postgresql"
DB_NAME="protei_bulk"
DB_USER="protei_user"
DB_HOST="localhost"

BACKUP_FILE="$BACKUP_DIR/protei_bulk_${TIMESTAMP}.sql"

echo "Database: $DB_NAME"
echo "Backup file: $BACKUP_FILE"

# Perform backup based on database type
case $DB_TYPE in
    postgresql)
        echo "Backing up PostgreSQL database..."
        pg_dump -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" > "$BACKUP_FILE"
        ;;
    mysql)
        echo "Backing up MySQL database..."
        mysqldump -h "$DB_HOST" -u "$DB_USER" -p "$DB_NAME" > "$BACKUP_FILE"
        ;;
    *)
        echo -e "${RED}Error: Unsupported database type: $DB_TYPE${NC}"
        exit 1
        ;;
esac

if [ $? -eq 0 ]; then
    # Compress backup
    gzip "$BACKUP_FILE"
    echo -e "${GREEN}✓ Backup completed successfully${NC}"
    echo "Backup: ${BACKUP_FILE}.gz"
    echo "Size: $(du -h "${BACKUP_FILE}.gz" | cut -f1)"
else
    echo -e "${RED}✗ Backup failed${NC}"
    exit 1
fi
