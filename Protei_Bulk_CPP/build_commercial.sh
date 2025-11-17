#!/bin/bash
#
# Commercial Build Script for Protei_Bulk Enterprise Edition
# This script creates a protected, optimized, single-file deployment package
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                          ║${NC}"
echo -e "${GREEN}║     PROTEI_BULK ENTERPRISE EDITION                       ║${NC}"
echo -e "${GREEN}║     Commercial Build System                              ║${NC}"
echo -e "${GREEN}║                                                          ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build_commercial"
RELEASE_DIR="${PROJECT_ROOT}/release"
VERSION="1.0.0"
BUILD_ID="$(date +%Y%m%d%H%M%S)"

echo -e "${BLUE}[1/10]${NC} Cleaning previous builds..."
rm -rf "${BUILD_DIR}" "${RELEASE_DIR}"
mkdir -p "${BUILD_DIR}" "${RELEASE_DIR}"

echo -e "${BLUE}[2/10]${NC} Configuring CMake for production..."
cd "${BUILD_DIR}"
cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_FLAGS="-O3 -DNDEBUG -march=native -mtune=native -flto" \
    -DCMAKE_EXE_LINKER_FLAGS="-Wl,--strip-all -Wl,--gc-sections" \
    -DCMAKE_INSTALL_PREFIX="/opt/protei_bulk"

echo -e "${BLUE}[3/10]${NC} Building with maximum optimization..."
make -j$(nproc) VERBOSE=0

echo -e "${BLUE}[4/10]${NC} Stripping binary and removing debug symbols..."
strip --strip-all bin/protei_bulk

echo -e "${BLUE}[5/10]${NC} Compressing binary with UPX..."
if command -v upx &> /dev/null; then
    upx --best --lzma bin/protei_bulk 2>/dev/null || echo "UPX compression skipped"
else
    echo -e "${YELLOW}UPX not installed, skipping compression${NC}"
    echo "Install with: sudo apt-get install upx-ucl"
fi

echo -e "${BLUE}[6/10]${NC} Creating deployment package..."
mkdir -p "${RELEASE_DIR}/protei_bulk_v${VERSION}"
cp bin/protei_bulk "${RELEASE_DIR}/protei_bulk_v${VERSION}/"

# Copy configuration templates
mkdir -p "${RELEASE_DIR}/protei_bulk_v${VERSION}/config"
cp -r "${PROJECT_ROOT}/config/"* "${RELEASE_DIR}/protei_bulk_v${VERSION}/config/"

# Create directory structure
cd "${RELEASE_DIR}/protei_bulk_v${VERSION}"
mkdir -p {logs,cdr,data,backup}

echo -e "${BLUE}[7/10]${NC} Creating installation script..."
cat > install.sh << 'INSTALL_SCRIPT_EOF'
#!/bin/bash
# Protei_Bulk Enterprise Edition - Installation Script

set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  Protei_Bulk Enterprise Edition                          ║"
echo "║  Installer v1.0.0                                        ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./install.sh)"
    exit 1
fi

# Install location
INSTALL_DIR="/opt/protei_bulk"

echo "[1/6] Creating installation directory..."
mkdir -p ${INSTALL_DIR}/{bin,config,logs,cdr,data,backup}

echo "[2/6] Copying files..."
cp protei_bulk ${INSTALL_DIR}/bin/
cp -r config/* ${INSTALL_DIR}/config/
chmod +x ${INSTALL_DIR}/bin/protei_bulk

echo "[3/6] Creating system user..."
if ! id -u protei &>/dev/null; then
    useradd -r -s /bin/false -d ${INSTALL_DIR} protei
fi

echo "[4/6] Setting permissions..."
chown -R protei:protei ${INSTALL_DIR}
chmod 750 ${INSTALL_DIR}/bin/protei_bulk
chmod 640 ${INSTALL_DIR}/config/*

echo "[5/6] Creating systemd service..."
cat > /etc/systemd/system/protei_bulk.service << 'SERVICE_EOF'
[Unit]
Description=Protei_Bulk Enterprise Messaging Platform
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=protei
Group=protei
WorkingDirectory=/opt/protei_bulk
ExecStart=/opt/protei_bulk/bin/protei_bulk
Restart=always
RestartSec=10
LimitNOFILE=65536
StandardOutput=journal
StandardError=journal

# Environment variables
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=protei_bulk"
Environment="DB_USER=protei"
Environment="DB_PASSWORD=elephant"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"

[Install]
WantedBy=multi-user.target
SERVICE_EOF

systemctl daemon-reload

echo "[6/6] Installation complete!"
echo
echo "╔══════════════════════════════════════════════════════════╗"
echo "║  Installation Successful                                 ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo
echo "Next steps:"
echo "1. Configure database connection in /opt/protei_bulk/config/db.conf"
echo "2. Start service: systemctl start protei_bulk"
echo "3. Enable on boot: systemctl enable protei_bulk"
echo "4. Check status: systemctl status protei_bulk"
echo "5. View logs: journalctl -u protei_bulk -f"
echo
echo "Access Points:"
echo "  API:  http://localhost:8080/api/v1"
echo "  SMPP: localhost:2775"
echo "  Logs: /opt/protei_bulk/logs/"
echo

INSTALL_SCRIPT_EOF

chmod +x install.sh

echo -e "${BLUE}[8/10]${NC} Creating uninstall script..."
cat > uninstall.sh << 'UNINSTALL_SCRIPT_EOF'
#!/bin/bash
# Protei_Bulk Enterprise Edition - Uninstall Script

set -e

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./uninstall.sh)"
    exit 1
fi

echo "Stopping service..."
systemctl stop protei_bulk 2>/dev/null || true
systemctl disable protei_bulk 2>/dev/null || true

echo "Removing service file..."
rm -f /etc/systemd/system/protei_bulk.service
systemctl daemon-reload

echo "Backing up configuration and data..."
BACKUP_DIR="/opt/protei_bulk_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p ${BACKUP_DIR}
cp -r /opt/protei_bulk/config ${BACKUP_DIR}/
cp -r /opt/protei_bulk/logs ${BACKUP_DIR}/
cp -r /opt/protei_bulk/cdr ${BACKUP_DIR}/

echo "Removing installation..."
rm -rf /opt/protei_bulk

echo "Uninstall complete! Backup saved to: ${BACKUP_DIR}"

UNINSTALL_SCRIPT_EOF

chmod +x uninstall.sh

echo -e "${BLUE}[9/10]${NC} Creating documentation..."
cat > README.txt << 'README_EOF'
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║     PROTEI_BULK ENTERPRISE EDITION v1.0.0                ║
║     High-Performance Bulk Messaging Platform             ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝

PRODUCT OVERVIEW
================
Protei_Bulk is an enterprise-grade bulk messaging platform designed
for telecom operators and service providers. It provides high-
performance message routing, campaign management, and multi-channel
messaging capabilities.

KEY FEATURES
============
✓ 15,000+ TPS sustained throughput
✓ Multi-channel support (SMS, WhatsApp, Email, Viber, RCS, Voice)
✓ SMPP 3.3/3.4/5.0 protocol support
✓ Advanced routing engine with 7 condition types
✓ Campaign management and scheduling
✓ Real-time analytics and reporting
✓ Subscriber profiling and segmentation
✓ GDPR/PDPL compliance
✓ Multi-tenancy support
✓ Enterprise-grade security

SYSTEM REQUIREMENTS
===================
- OS: Linux (Ubuntu 20.04+, CentOS 8+, Debian 11+)
- CPU: Minimum 4 cores, Recommended 16+ cores
- RAM: Minimum 8GB, Recommended 32GB+
- Disk: 100GB+ SSD
- Network: 1Gbps+

DEPENDENCIES
============
- PostgreSQL 14+
- Redis 7.0+
- Boost Libraries 1.75+
- OpenSSL 1.1+

INSTALLATION
============
1. Extract the package:
   tar -xzf protei_bulk_v1.0.0.tar.gz
   cd protei_bulk_v1.0.0

2. Run installation:
   sudo ./install.sh

3. Configure database:
   Edit /opt/protei_bulk/config/db.conf

4. Start service:
   sudo systemctl start protei_bulk

5. Check status:
   sudo systemctl status protei_bulk

LOGS
====
All logs are stored in /opt/protei_bulk/logs/:

- application.log  : General application logs
- warning.log      : Warnings and non-critical issues
- alarm.log        : Critical errors and system alarms
- system.log       : Performance metrics (CPU, memory, disk)
- cdr.log          : Call Detail Records
- security.log     : Security events

CDR FORMAT
==========
CDR files contain complete message statistics in CSV format:
message_id, campaign_id, customer_id, msisdn, sender_id,
message_text, length, parts, submit_time, delivery_time,
status, error_code, smsc_id, route_id, cost, operator,
country_code, retry_count, final_status, processing_time_ms

MONITORING
==========
System metrics are logged every 60 seconds to system.log:
- CPU usage percentage
- Memory usage (MB and %)
- Disk usage and available space
- Active connections
- Queue depth
- Messages per second

ALARMS
======
Critical events are logged to alarm.log:
- CPU usage > 90%
- Memory usage > 85%
- Disk space < 1GB
- Queue depth > 10,000 messages
- Database connection failures
- SMPP connection issues

API ENDPOINTS
=============
Base URL: http://localhost:8080/api/v1

Authentication:
  POST /auth/login

Messages:
  POST /messages/send
  GET  /messages/{id}

Campaigns:
  GET  /campaigns
  POST /campaigns
  POST /campaigns/{id}/start

See full API documentation at: http://localhost:8080/api/docs

SUPPORT
=======
For technical support and inquiries:
Email: support@protei-bulk.com
Documentation: https://docs.protei-bulk.com

LICENSE
=======
This is proprietary commercial software.
All rights reserved © 2025 Protei Systems

Unauthorized copying, modification, or distribution is
strictly prohibited and will be prosecuted.

README_EOF

echo -e "${BLUE}[10/10]${NC} Creating deployment archive..."
cd "${RELEASE_DIR}"
tar -czf "protei_bulk_enterprise_v${VERSION}_build${BUILD_ID}.tar.gz" "protei_bulk_v${VERSION}"

# Create checksum
sha256sum "protei_bulk_enterprise_v${VERSION}_build${BUILD_ID}.tar.gz" > "protei_bulk_enterprise_v${VERSION}_build${BUILD_ID}.tar.gz.sha256"

echo
echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Commercial Build Complete!                              ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo
echo -e "${GREEN}Package Information:${NC}"
echo -e "  Version: ${VERSION}"
echo -e "  Build ID: ${BUILD_ID}"
echo -e "  Location: ${RELEASE_DIR}"
echo
ls -lh "${RELEASE_DIR}/"*.tar.gz
echo
echo -e "${YELLOW}Binary Size:${NC}"
ls -lh "${BUILD_DIR}/bin/protei_bulk"
echo
echo -e "${GREEN}Ready for deployment!${NC}"
echo
