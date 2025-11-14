#!/bin/bash
#
# Protei Monitoring v2.0 - Deployment Script
#
# This script deploys Protei Monitoring to a production server
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VERSION="2.0.0"
PACKAGE_NAME="Protei_Monitoring-${VERSION}"

print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring v${VERSION} - Deployment"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

print_header

# Step 1: Preparation
print_step "Preparing deployment..."

DEPLOY_DIR="/usr/protei"
TARGET_DIR="$DEPLOY_DIR/Protei_Monitoring"
BACKUP_DIR="$DEPLOY_DIR/backup/$(date +%Y%m%d_%H%M%S)"

# Step 2: Backup existing installation
if [ -d "$TARGET_DIR" ]; then
    print_step "Backing up existing installation..."
    mkdir -p "$BACKUP_DIR"

    # Backup configuration
    if [ -d "$TARGET_DIR/config" ]; then
        cp -r "$TARGET_DIR/config" "$BACKUP_DIR/"
        print_success "Configuration backed up to $BACKUP_DIR"
    fi

    # Backup logs (last 7 days)
    if [ -d "$TARGET_DIR/logs" ]; then
        find "$TARGET_DIR/logs" -name "*.log" -mtime -7 -exec cp --parents {} "$BACKUP_DIR/" \;
        print_success "Recent logs backed up"
    fi

    # Stop running application
    if [ -f "$TARGET_DIR/scripts/stop" ]; then
        print_step "Stopping running application..."
        "$TARGET_DIR/scripts/stop" || true
        sleep 2
    fi
fi

# Step 3: Deploy new version
print_step "Deploying Protei Monitoring v${VERSION}..."

mkdir -p "$DEPLOY_DIR"

# Copy package to target
if [ -d "/home/user/omar251990/Protei_Monitoring" ]; then
    # Development deployment
    print_info "Deploying from source directory..."
    rsync -av --delete /home/user/omar251990/Protei_Monitoring/ "$TARGET_DIR/"
else
    print_error "Source directory not found"
    exit 1
fi

print_success "Files deployed to $TARGET_DIR"

# Step 4: Restore configuration
if [ -d "$BACKUP_DIR/config" ]; then
    print_step "Restoring configuration..."
    cp -r "$BACKUP_DIR/config/"* "$TARGET_DIR/config/"
    print_success "Configuration restored"
fi

# Step 5: Set permissions
print_step "Setting permissions..."

# Create protei user if doesn't exist
if ! id -u protei &>/dev/null; then
    useradd -r -s /bin/false protei
    print_success "Created protei user"
fi

chown -R root:protei "$TARGET_DIR"
chown -R protei:protei "$TARGET_DIR/logs"
chown -R protei:protei "$TARGET_DIR/cdr"
chown -R protei:protei "$TARGET_DIR/tmp"

chmod -R 750 "$TARGET_DIR/config"
chmod -R 755 "$TARGET_DIR/bin"
chmod -R 755 "$TARGET_DIR/scripts"
chmod -R 755 "$TARGET_DIR/logs"
chmod -R 755 "$TARGET_DIR/cdr"

chmod 600 "$TARGET_DIR/config/db.cfg"
chmod 600 "$TARGET_DIR/config/license.cfg"
chmod 600 "$TARGET_DIR/config/security.cfg"

chmod +x "$TARGET_DIR/scripts/"*
chmod +x "$TARGET_DIR/scripts/utils/"*

print_success "Permissions set"

# Step 6: Run installation
print_step "Running installation script..."
if [ -f "$TARGET_DIR/scripts/install.sh" ]; then
    "$TARGET_DIR/scripts/install.sh"
else
    print_info "Installation script not found, skipping..."
fi

# Step 7: Start application
print_step "Starting application..."
if [ -f "$TARGET_DIR/scripts/start" ]; then
    "$TARGET_DIR/scripts/start"
    sleep 3

    # Verify startup
    if [ -f "$TARGET_DIR/scripts/status" ]; then
        if "$TARGET_DIR/scripts/status" &>/dev/null; then
            print_success "Application started successfully"
        else
            print_error "Application failed to start. Check logs in $TARGET_DIR/logs/"
            exit 1
        fi
    fi
else
    print_info "Start script not found"
fi

# Step 8: Run tests
print_step "Running system tests..."
if [ -f "$TARGET_DIR/scripts/test.sh" ]; then
    if "$TARGET_DIR/scripts/test.sh"; then
        print_success "All tests passed"
    else
        print_error "Some tests failed. Check output above."
    fi
else
    print_info "Test script not found, skipping tests..."
fi

# Summary
echo ""
echo -e "${GREEN}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Deployment Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"
echo ""
echo "Version: $VERSION"
echo "Install Directory: $TARGET_DIR"
echo "Backup Directory: $BACKUP_DIR"
echo ""
echo "Next Steps:"
echo "  1. Verify application status: $TARGET_DIR/scripts/status"
echo "  2. Check logs: tail -f $TARGET_DIR/logs/application/protei-monitoring.log"
echo "  3. Access web interface: http://<server_ip>:8080"
echo "  4. Review deployment backup: $BACKUP_DIR"
echo ""
print_success "Protei Monitoring v${VERSION} deployed successfully!"

exit 0
