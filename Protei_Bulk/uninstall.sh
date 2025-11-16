#!/bin/bash
################################################################################
# Protei_Bulk - Uninstallation Script
# Removes Protei_Bulk and all associated components
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Settings
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_USER="protei"
DB_NAME="protei_bulk"
APP_USER="protei"
SERVICE_NAME="protei_bulk"

echo -e "${RED}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                                                                ║"
echo "║          Protei_Bulk Uninstallation Script                    ║"
echo "║                                                                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""
echo -e "${YELLOW}WARNING: This will remove:${NC}"
echo "  • Protei_Bulk application files"
echo "  • Database: $DB_NAME"
echo "  • Database user: $DB_USER"
echo "  • System user: $APP_USER"
echo "  • Systemd service: $SERVICE_NAME"
echo "  • Python virtual environment"
echo ""
echo -e "${RED}THIS CANNOT BE UNDONE!${NC}"
echo ""

read -p "Are you absolutely sure? Type 'yes' to confirm: " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Uninstallation cancelled."
    exit 0
fi

echo ""
echo "Starting uninstallation..."
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

# Stop and disable service
echo "Stopping Protei_Bulk service..."
systemctl stop $SERVICE_NAME 2>/dev/null || true
systemctl disable $SERVICE_NAME 2>/dev/null || true

# Remove systemd service file
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    echo "Removing systemd service..."
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
fi

# Drop database
echo "Removing database..."
sudo -u postgres psql -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true

# Drop database user
echo "Removing database user..."
sudo -u postgres psql -c "DROP USER IF EXISTS $DB_USER;" 2>/dev/null || true

# Remove application user
if id "$APP_USER" &>/dev/null; then
    echo "Removing application user..."
    userdel -r "$APP_USER" 2>/dev/null || true
fi

# Optional: Remove application files
echo ""
read -p "Remove application files from $INSTALL_DIR? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing application files..."
    cd /
    rm -rf "$INSTALL_DIR"
    echo "Application files removed."
else
    echo "Application files kept at: $INSTALL_DIR"
fi

echo ""
echo -e "${GREEN}✓ Uninstallation complete${NC}"
echo ""
