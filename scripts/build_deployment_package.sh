#!/bin/bash

################################################################################
# Protei Monitoring - Complete Deployment Package Builder
# This script:
#   1. Builds the application binary
#   2. Creates deployment package with all necessary files
#   3. Encrypts the package for secure distribution
#   4. Generates deployment instructions
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"
OUTPUT_DIR="$PROJECT_ROOT/dist"
TEMP_BUILD_DIR="/tmp/protei-build-$$"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}========================================${NC}\n"
}

# Check dependencies
check_dependencies() {
    print_header "Checking Dependencies"

    local missing_deps=()

    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi

    if ! command -v openssl &> /dev/null; then
        missing_deps+=("openssl")
    fi

    if ! command -v tar &> /dev/null; then
        missing_deps+=("tar")
    fi

    if ! command -v git &> /dev/null; then
        print_warning "git not found - version info will be limited"
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_info "Please install missing dependencies and try again"
        exit 1
    fi

    print_success "All required dependencies found"
}

# Clean previous builds
clean_builds() {
    print_header "Cleaning Previous Builds"

    rm -rf $OUTPUT_DIR
    rm -rf $TEMP_BUILD_DIR

    mkdir -p $OUTPUT_DIR
    mkdir -p $TEMP_BUILD_DIR

    print_success "Clean completed"
}

# Build application
build_application() {
    print_header "Building Application Binary"

    cd $PROJECT_ROOT

    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Not a Go project?"
        exit 1
    fi

    print_info "Downloading Go modules..."
    go mod download

    # Get version info
    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        VERSION=$(git describe --tags --always 2>/dev/null || echo "2.0.0")
        GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    else
        VERSION="2.0.0"
        GIT_COMMIT="unknown"
    fi

    BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    print_info "Version: $VERSION"
    print_info "Commit: ${GIT_COMMIT:0:8}"
    print_info "Build Time: $BUILD_TIME"

    # Create bin directory
    mkdir -p $TEMP_BUILD_DIR/bin

    print_info "Compiling binary with optimizations..."

    # Build with optimizations
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT -X main.BuildTime=$BUILD_TIME" \
        -trimpath \
        -o $TEMP_BUILD_DIR/bin/protei-monitoring \
        ./cmd/protei-monitoring

    if [ ! -f "$TEMP_BUILD_DIR/bin/protei-monitoring" ]; then
        print_error "Build failed - binary not created"
        exit 1
    fi

    # Strip debugging symbols
    strip $TEMP_BUILD_DIR/bin/protei-monitoring 2>/dev/null || true

    # Get binary info
    BINARY_SIZE=$(du -h $TEMP_BUILD_DIR/bin/protei-monitoring | awk '{print $1}')
    print_success "Binary built successfully"
    print_info "Binary size: $BINARY_SIZE"
}

# Copy application files
copy_application_files() {
    print_header "Preparing Application Files"

    # Create directory structure
    mkdir -p $TEMP_BUILD_DIR/config
    mkdir -p $TEMP_BUILD_DIR/scripts
    mkdir -p $TEMP_BUILD_DIR/docs
    mkdir -p $TEMP_BUILD_DIR/web/static
    mkdir -p $TEMP_BUILD_DIR/web/templates

    # Copy web assets if they exist
    if [ -d "$PROJECT_ROOT/web/static" ]; then
        cp -r $PROJECT_ROOT/web/static/* $TEMP_BUILD_DIR/web/static/ 2>/dev/null || true
        print_info "Copied web static files"
    fi

    if [ -d "$PROJECT_ROOT/web/templates" ]; then
        cp -r $PROJECT_ROOT/web/templates/* $TEMP_BUILD_DIR/web/templates/ 2>/dev/null || true
        print_info "Copied web templates"
    fi

    # Copy configuration templates
    if [ -f "$PROJECT_ROOT/config/config.yaml.example" ]; then
        cp $PROJECT_ROOT/config/config.yaml.example $TEMP_BUILD_DIR/config/
    else
        # Create example config
        cat > $TEMP_BUILD_DIR/config/config.yaml.example <<'EOF'
# Protei Monitoring Configuration Example

server:
  host: "0.0.0.0"
  port: 8443
  tls:
    enabled: true
    cert_file: "/etc/protei-monitoring/certs/server.crt"
    key_file: "/etc/protei-monitoring/certs/server.key"

database:
  host: "localhost"
  port: 5432
  database: "protei_monitoring"
  username: "protei_user"
  password: "CHANGE_ME"
  ssl_mode: "disable"
  max_connections: 100

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

capture:
  interfaces:
    - "any"
  pcap_dir: "/var/lib/protei-monitoring/pcap"
  buffer_size: 67108864
  max_file_size: 1073741824

logging:
  level: "info"
  file: "/var/log/protei-monitoring/protei-monitoring.log"
  max_size: 100
  max_backups: 10
  max_age: 30

license:
  key: "YOUR_LICENSE_KEY"
  company: "Your Company"
  max_users: 100
EOF
    fi

    # Copy installation script
    cp $SCRIPTS_DIR/install.sh $TEMP_BUILD_DIR/scripts/
    chmod +x $TEMP_BUILD_DIR/scripts/install.sh

    print_success "Application files prepared"
}

# Create version file
create_version_file() {
    print_header "Creating Version Information"

    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        VERSION=$(git describe --tags --always 2>/dev/null || echo "2.0.0")
        GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    else
        VERSION="2.0.0"
        GIT_COMMIT="unknown"
    fi

    cat > $TEMP_BUILD_DIR/VERSION <<EOF
Product: Protei Monitoring
Version: $VERSION
Build Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Git Commit: $GIT_COMMIT
Build Host: $(hostname)
Go Version: $(go version | awk '{print $3}')
Platform: linux/amd64
EOF

    cat > $TEMP_BUILD_DIR/README.md <<'EOF'
# Protei Monitoring - Enterprise Telecom Protocol Monitoring System

## Quick Start

For automated installation on a Linux server:

```bash
sudo bash scripts/install.sh
```

This will automatically:
- Detect your OS (Ubuntu, Debian, CentOS, RHEL, Rocky Linux, Fedora)
- Install all dependencies (Go, PostgreSQL, Redis)
- Configure database and services
- Set up the application
- Generate SSL certificates
- Create systemd service
- Configure firewall

## Manual Installation

See `docs/INSTALLATION.md` for detailed manual installation instructions.

## Features

- **Multi-Protocol Support**: MAP, CAP, INAP, Diameter, GTP, PFCP, NGAP, S1AP, NAS
- **AI-Based Analysis**: Automatic issue detection and recommendations
- **Flow Reconstruction**: Reconstruct signaling flows per 3GPP procedures
- **Subscriber Correlation**: Track subscribers across all interfaces
- **3GPP Knowledge Base**: Built-in protocol and standards reference
- **Real-time Monitoring**: Live traffic capture and analysis
- **Web Interface**: Modern web-based management interface

## Access

After installation:
- Web Interface: https://YOUR_IP:8443
- Admin credentials: `/etc/protei-monitoring/admin_credentials.txt`

## System Requirements

- Linux OS (Ubuntu 20.04+, CentOS 8+, Debian 11+, Rocky Linux 8+)
- 2+ CPU cores
- 4+ GB RAM
- 50+ GB disk space
- Root or sudo access

## Support

For support and licensing:
- Email: support@protei-monitoring.com
- Documentation: https://docs.protei-monitoring.com

## License

Copyright © 2025 Protei Monitoring. All rights reserved.
Proprietary and confidential software.
EOF

    print_success "Version information created"
}

# Generate encryption key
generate_encryption_key() {
    print_header "Generating Encryption Key"

    echo ""
    read -p "Do you want to (1) Generate new key or (2) Use existing key? [1/2]: " choice
    echo ""

    if [ "$choice" = "2" ]; then
        read -sp "Enter your encryption key: " ENCRYPT_KEY
        echo ""
    else
        # Generate strong random key
        ENCRYPT_KEY=$(openssl rand -base64 32)

        print_warning "A strong encryption key has been generated"
        print_error "SAVE THIS KEY - You cannot recover the package without it!"
        echo ""
        echo -e "${YELLOW}Encryption Key:${NC} ${RED}$ENCRYPT_KEY${NC}"
        echo ""
        read -p "Press ENTER after you have safely stored this key..."
        echo ""

        # Save to file
        echo "$ENCRYPT_KEY" > $OUTPUT_DIR/ENCRYPTION_KEY.txt
        chmod 600 $OUTPUT_DIR/ENCRYPTION_KEY.txt
    fi
}

# Create encrypted deployment package
create_encrypted_package() {
    print_header "Creating Encrypted Deployment Package"

    # Create tarball
    print_info "Creating tarball..."
    cd $TEMP_BUILD_DIR
    tar -czf protei-monitoring.tar.gz *

    # Encrypt
    print_info "Encrypting with AES-256-CBC..."
    openssl enc -aes-256-cbc -salt -pbkdf2 -iter 100000 \
        -in protei-monitoring.tar.gz \
        -out $OUTPUT_DIR/protei-monitoring-${TIMESTAMP}.enc \
        -k "$ENCRYPT_KEY"

    if [ ! -f "$OUTPUT_DIR/protei-monitoring-${TIMESTAMP}.enc" ]; then
        print_error "Encryption failed"
        exit 1
    fi

    # Generate checksum
    cd $OUTPUT_DIR
    sha256sum protei-monitoring-${TIMESTAMP}.enc > protei-monitoring-${TIMESTAMP}.enc.sha256

    PACKAGE_SIZE=$(du -h protei-monitoring-${TIMESTAMP}.enc | awk '{print $1}')
    CHECKSUM=$(cat protei-monitoring-${TIMESTAMP}.enc.sha256 | awk '{print $1}')

    print_success "Encrypted package created"
    print_info "Size: $PACKAGE_SIZE"
    print_info "SHA256: ${CHECKSUM:0:32}..."
}

# Create deployment instructions
create_deployment_instructions() {
    print_header "Creating Deployment Instructions"

    cat > $OUTPUT_DIR/DEPLOY.md <<EOF
# Protei Monitoring - Deployment Instructions
Generated: $(date)

## Package Information

- **Package**: protei-monitoring-${TIMESTAMP}.enc
- **Size**: $(du -h $OUTPUT_DIR/protei-monitoring-${TIMESTAMP}.enc | awk '{print $1}')
- **SHA256**: $(cat $OUTPUT_DIR/protei-monitoring-${TIMESTAMP}.enc.sha256 | awk '{print $1}')
- **Build Date**: $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Important Security Information

### Encryption Key

The package is encrypted with AES-256-CBC encryption. You need the encryption key to:
1. Install the application on a server
2. Update the application
3. Access the deployment files

**CRITICAL**: Keep your encryption key secure!
- Store in a password manager
- Create backup copies in secure locations
- Never commit to version control
- Never share via unsecured channels

### Decryption Process

The installation script will automatically decrypt the package when you provide the key.

## Prerequisites

Target server requirements:
- **OS**: Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+, Rocky Linux 8+, Fedora 34+
- **CPU**: 2+ cores
- **RAM**: 4+ GB minimum, 8+ GB recommended
- **Disk**: 50+ GB minimum
- **Access**: Root or sudo privileges
- **Network**: Internet connection for dependency installation

## Automated Installation (Recommended)

### Step 1: Transfer Package to Server

\`\`\`bash
# From your local machine
scp protei-monitoring-${TIMESTAMP}.enc root@YOUR_SERVER:/tmp/
\`\`\`

### Step 2: Connect to Server

\`\`\`bash
ssh root@YOUR_SERVER
\`\`\`

### Step 3: Decrypt and Extract

\`\`\`bash
cd /tmp

# Decrypt (you'll be prompted for the encryption key)
openssl enc -aes-256-cbc -d -pbkdf2 -iter 100000 \\
    -in protei-monitoring-${TIMESTAMP}.enc \\
    -out protei-monitoring.tar.gz \\
    -k YOUR_ENCRYPTION_KEY

# Verify checksum (optional but recommended)
sha256sum protei-monitoring.tar.gz

# Extract
tar -xzf protei-monitoring.tar.gz
\`\`\`

### Step 4: Run Installation

\`\`\`bash
cd scripts
chmod +x install.sh
sudo ./install.sh
\`\`\`

The installer will:
1. Detect your operating system
2. Update system packages
3. Install dependencies (Go, PostgreSQL, Redis)
4. Create application user and directories
5. Setup and configure PostgreSQL database
6. Deploy the application
7. Generate SSL certificates
8. Create and start systemd service
9. Configure firewall rules
10. Create admin user credentials

Installation typically takes 5-10 minutes depending on your server.

### Step 5: Access the Application

After installation completes:

1. **Web Interface**: https://YOUR_SERVER_IP:8443
2. **Admin Credentials**: \`/etc/protei-monitoring/admin_credentials.txt\`
3. **Check Status**: \`systemctl status protei-monitoring\`

## Manual Installation

If you need to install manually or customize the installation:

### 1. Install Dependencies

**Ubuntu/Debian:**
\`\`\`bash
apt update
apt install -y postgresql-14 postgresql-contrib redis-server \\
    build-essential libpcap-dev tcpdump git wget curl
\`\`\`

**CentOS/RHEL/Rocky:**
\`\`\`bash
dnf install -y postgresql-server postgresql-contrib redis \\
    gcc gcc-c++ libpcap-devel tcpdump git wget curl
\`\`\`

### 2. Install Go 1.21+

\`\`\`bash
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
\`\`\`

### 3. Setup PostgreSQL

\`\`\`bash
# Initialize database (CentOS/RHEL)
postgresql-setup --initdb  # Skip on Ubuntu/Debian

# Start service
systemctl start postgresql
systemctl enable postgresql

# Create database and user
sudo -u postgres psql <<EOSQL
CREATE USER protei_user WITH PASSWORD 'YOUR_SECURE_PASSWORD';
CREATE DATABASE protei_monitoring OWNER protei_user;
GRANT ALL PRIVILEGES ON DATABASE protei_monitoring TO protei_user;
\\c protei_monitoring
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
EOSQL
\`\`\`

### 4. Setup Redis

\`\`\`bash
systemctl start redis
systemctl enable redis
\`\`\`

### 5. Deploy Application

\`\`\`bash
# Create directories
mkdir -p /opt/protei-monitoring/bin
mkdir -p /etc/protei-monitoring
mkdir -p /var/log/protei-monitoring
mkdir -p /var/lib/protei-monitoring

# Create user
useradd -r -s /bin/bash -d /opt/protei-monitoring protei

# Copy files from extracted package
cp bin/protei-monitoring /opt/protei-monitoring/bin/
cp -r web /opt/protei-monitoring/
cp config/config.yaml.example /etc/protei-monitoring/config.yaml

# Edit configuration
vim /etc/protei-monitoring/config.yaml
# Update database password and other settings

# Set ownership
chown -R protei:protei /opt/protei-monitoring
chown -R protei:protei /etc/protei-monitoring
chown -R protei:protei /var/log/protei-monitoring
chown -R protei:protei /var/lib/protei-monitoring

# Set capabilities for packet capture
setcap 'cap_net_raw,cap_net_admin+eip' /opt/protei-monitoring/bin/protei-monitoring
\`\`\`

### 6. Create Systemd Service

Create \`/etc/systemd/system/protei-monitoring.service\`:

\`\`\`ini
[Unit]
Description=Protei Monitoring - Telecom Protocol Monitoring
After=network.target postgresql.service redis.service
Requires=postgresql.service redis.service

[Service]
Type=simple
User=protei
Group=protei
WorkingDirectory=/opt/protei-monitoring
ExecStart=/opt/protei-monitoring/bin/protei-monitoring -config /etc/protei-monitoring/config.yaml
Restart=always
RestartSec=10
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
\`\`\`

### 7. Start Service

\`\`\`bash
systemctl daemon-reload
systemctl enable protei-monitoring
systemctl start protei-monitoring
systemctl status protei-monitoring
\`\`\`

## Post-Installation

### 1. Change Default Password

Login to https://YOUR_IP:8443 with the admin credentials and change the password immediately.

### 2. Install Valid SSL Certificate

Replace the self-signed certificate:

\`\`\`bash
# Copy your certificates
cp your-cert.crt /etc/protei-monitoring/certs/server.crt
cp your-key.key /etc/protei-monitoring/certs/server.key

# Set permissions
chown protei:protei /etc/protei-monitoring/certs/*
chmod 600 /etc/protei-monitoring/certs/server.key

# Restart service
systemctl restart protei-monitoring
\`\`\`

### 3. Configure Firewall

\`\`\`bash
# firewalld (CentOS/RHEL/Rocky)
firewall-cmd --permanent --add-port=8443/tcp
firewall-cmd --reload

# ufw (Ubuntu/Debian)
ufw allow 8443/tcp
\`\`\`

### 4. Setup Backup

\`\`\`bash
# Backup script example
#!/bin/bash
BACKUP_DIR="/backup/protei-monitoring"
DATE=\$(date +%Y%m%d_%H%M%S)

# Backup database
sudo -u postgres pg_dump protei_monitoring > \$BACKUP_DIR/db_\$DATE.sql

# Backup data directory
tar -czf \$BACKUP_DIR/data_\$DATE.tar.gz /var/lib/protei-monitoring

# Backup configuration
tar -czf \$BACKUP_DIR/config_\$DATE.tar.gz /etc/protei-monitoring
\`\`\`

## Troubleshooting

### Service Won't Start

\`\`\`bash
# Check status
systemctl status protei-monitoring

# View logs
journalctl -u protei-monitoring -n 100 -f

# Check configuration
/opt/protei-monitoring/bin/protei-monitoring -config /etc/protei-monitoring/config.yaml -check-config
\`\`\`

### Database Connection Issues

\`\`\`bash
# Test connection
sudo -u postgres psql -d protei_monitoring -U protei_user

# Check PostgreSQL is running
systemctl status postgresql

# Check PostgreSQL logs
tail -f /var/log/postgresql/postgresql-14-main.log
\`\`\`

### Cannot Capture Packets

\`\`\`bash
# Verify capabilities
getcap /opt/protei-monitoring/bin/protei-monitoring
# Should show: cap_net_raw,cap_net_admin+eip

# Re-apply if missing
setcap 'cap_net_raw,cap_net_admin+eip' /opt/protei-monitoring/bin/protei-monitoring
\`\`\`

### Web Interface Not Accessible

\`\`\`bash
# Check if service is listening
netstat -tlnp | grep 8443

# Check firewall
firewall-cmd --list-ports  # CentOS/RHEL
ufw status                 # Ubuntu

# Check SSL certificates
openssl s_client -connect localhost:8443
\`\`\`

## Updating

To update to a new version:

1. Stop the service: \`systemctl stop protei-monitoring\`
2. Backup current installation
3. Decrypt new package with your encryption key
4. Extract and copy new binary
5. Update configuration if needed
6. Start service: \`systemctl start protei-monitoring\`

## Uninstallation

\`\`\`bash
# Stop service
systemctl stop protei-monitoring
systemctl disable protei-monitoring

# Remove service file
rm /etc/systemd/system/protei-monitoring.service
systemctl daemon-reload

# Remove application files
rm -rf /opt/protei-monitoring
rm -rf /etc/protei-monitoring
rm -rf /var/log/protei-monitoring
rm -rf /var/lib/protei-monitoring

# Remove user
userdel protei

# Remove database (if desired)
sudo -u postgres psql -c "DROP DATABASE protei_monitoring;"
sudo -u postgres psql -c "DROP USER protei_user;"
\`\`\`

## Support and Contact

For technical support, licensing, or questions:
- **Email**: support@protei-monitoring.com
- **Documentation**: https://docs.protei-monitoring.com
- **Sales**: sales@protei-monitoring.com

## Security Best Practices

1. Change default admin password immediately
2. Use valid SSL certificates in production
3. Restrict network access to necessary ports only
4. Enable firewall rules
5. Regular backups of database and configuration
6. Keep encryption key secure and backed up
7. Monitor logs for suspicious activity
8. Regular security updates
9. Use strong database passwords
10. Enable audit logging

## License

Copyright © 2025 Protei Monitoring. All rights reserved.

This software is proprietary and confidential. Unauthorized copying, distribution,
or use of this software is strictly prohibited and may result in legal action.

All intellectual property rights, including but not limited to copyrights, patents,
and trade secrets, are owned by Protei Monitoring.

For licensing inquiries, contact: sales@protei-monitoring.com
EOF

    print_success "Deployment instructions created"
}

# Create quick start guide
create_quick_start() {
    print_header "Creating Quick Start Guide"

    cat > $OUTPUT_DIR/QUICKSTART.txt <<EOF
╔════════════════════════════════════════════════════════════════╗
║          PROTEI MONITORING - QUICK START GUIDE                 ║
╚════════════════════════════════════════════════════════════════╝

PACKAGE: protei-monitoring-${TIMESTAMP}.enc

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. TRANSFER TO SERVER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

scp protei-monitoring-${TIMESTAMP}.enc root@YOUR_SERVER:/tmp/

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. DECRYPT ON SERVER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ssh root@YOUR_SERVER
cd /tmp

openssl enc -aes-256-cbc -d -pbkdf2 -iter 100000 \\
    -in protei-monitoring-${TIMESTAMP}.enc \\
    -out protei-monitoring.tar.gz \\
    -k YOUR_ENCRYPTION_KEY

tar -xzf protei-monitoring.tar.gz

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. INSTALL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

cd scripts
chmod +x install.sh
sudo ./install.sh

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. ACCESS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Web Interface: https://YOUR_SERVER_IP:8443
Credentials:   /etc/protei-monitoring/admin_credentials.txt

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. MANAGE SERVICE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

systemctl status protei-monitoring
systemctl restart protei-monitoring
journalctl -u protei-monitoring -f

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  IMPORTANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Keep your encryption key SAFE and BACKED UP
2. Change admin password after first login
3. Replace self-signed SSL certificate for production
4. Configure firewall rules appropriately
5. Setup regular backups

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
For full documentation, see DEPLOY.md
Support: support@protei-monitoring.com
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EOF

    print_success "Quick start guide created"
}

# Cleanup
cleanup() {
    print_header "Cleaning Up"

    rm -rf $TEMP_BUILD_DIR

    print_success "Temporary files removed"
}

# Print summary
print_summary() {
    print_header "Build Complete!"

    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║    DEPLOYMENT PACKAGE CREATED SUCCESSFULLY             ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""

    echo -e "${YELLOW}Output Directory:${NC}"
    echo -e "  $OUTPUT_DIR"
    echo ""

    echo -e "${YELLOW}Generated Files:${NC}"
    ls -lh $OUTPUT_DIR | tail -n +2 | awk '{printf "  %-40s %8s\n", $9, $5}'
    echo ""

    echo -e "${RED}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  CRITICAL - SAVE YOUR ENCRYPTION KEY                   ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════╝${NC}"
    if [ -f "$OUTPUT_DIR/ENCRYPTION_KEY.txt" ]; then
        echo -e "${YELLOW}Location:${NC} $OUTPUT_DIR/ENCRYPTION_KEY.txt"
        echo -e "${YELLOW}Key:${NC} ${RED}$(cat $OUTPUT_DIR/ENCRYPTION_KEY.txt)${NC}"
    fi
    echo ""

    echo -e "${YELLOW}Next Steps:${NC}"
    echo -e "  1. ${GREEN}BACKUP${NC} your encryption key (store in password manager)"
    echo -e "  2. Read ${YELLOW}QUICKSTART.txt${NC} for deployment instructions"
    echo -e "  3. Transfer encrypted package to your server"
    echo -e "  4. Run installation script on server"
    echo ""

    print_success "Package ready for deployment!"
}

# Main execution
main() {
    print_header "Protei Monitoring - Deployment Package Builder"

    check_dependencies
    clean_builds
    build_application
    copy_application_files
    create_version_file
    generate_encryption_key
    create_encrypted_package
    create_deployment_instructions
    create_quick_start
    cleanup
    print_summary
}

# Run
main "$@"
