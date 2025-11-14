#!/bin/bash

################################################################################
# Protei Monitoring - Automated Installation Script
# This script automatically installs Protei Monitoring on Linux servers
# Supports: Ubuntu/Debian, CentOS/RHEL/Rocky Linux, Fedora
################################################################################

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="protei-monitoring"
APP_USER="protei"
APP_GROUP="protei"
INSTALL_DIR="/opt/protei-monitoring"
CONFIG_DIR="/etc/protei-monitoring"
LOG_DIR="/var/log/protei-monitoring"
DATA_DIR="/var/lib/protei-monitoring"
GO_VERSION="1.21.5"
POSTGRES_VERSION="14"
REDIS_VERSION="7"

# Print functions
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

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "Please run this script as root (use sudo)"
        exit 1
    fi
    print_success "Running as root"
}

# Detect OS
detect_os() {
    print_header "Detecting Operating System"

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
        OS_NAME=$PRETTY_NAME
    else
        print_error "Cannot detect OS. /etc/os-release not found"
        exit 1
    fi

    print_info "OS: $OS_NAME"
    print_info "Version: $OS_VERSION"

    case $OS in
        ubuntu|debian)
            PKG_MANAGER="apt"
            PKG_UPDATE="apt update"
            PKG_INSTALL="apt install -y"
            ;;
        centos|rhel|rocky|almalinux)
            PKG_MANAGER="yum"
            PKG_UPDATE="yum update -y"
            PKG_INSTALL="yum install -y"
            ;;
        fedora)
            PKG_MANAGER="dnf"
            PKG_UPDATE="dnf update -y"
            PKG_INSTALL="dnf install -y"
            ;;
        *)
            print_error "Unsupported OS: $OS"
            print_info "Supported: Ubuntu, Debian, CentOS, RHEL, Rocky Linux, AlmaLinux, Fedora"
            exit 1
            ;;
    esac

    print_success "OS detected: $OS ($PKG_MANAGER)"
}

# Update system packages
update_system() {
    print_header "Updating System Packages"
    print_info "This may take a few minutes..."

    $PKG_UPDATE

    print_success "System packages updated"
}

# Install basic dependencies
install_dependencies() {
    print_header "Installing Basic Dependencies"

    local deps="wget curl git vim htop net-tools tcpdump wireshark-common libpcap-dev gcc make openssl"

    if [ "$PKG_MANAGER" = "apt" ]; then
        deps="$deps build-essential"
        $PKG_INSTALL $deps
    else
        deps="$deps gcc-c++ kernel-devel"
        $PKG_INSTALL $deps
    fi

    print_success "Basic dependencies installed"
}

# Install Go
install_go() {
    print_header "Installing Go $GO_VERSION"

    # Check if Go is already installed
    if command -v go &> /dev/null; then
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go $CURRENT_GO is already installed"

        if [ "$CURRENT_GO" = "$GO_VERSION" ]; then
            print_success "Go version matches required version"
            return
        else
            print_warning "Different Go version detected. Installing required version..."
        fi
    fi

    # Download and install Go
    cd /tmp
    wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz

    # Remove old installation
    rm -rf /usr/local/go

    # Extract new version
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

    # Set up environment
    cat > /etc/profile.d/go.sh <<EOF
export GOROOT=/usr/local/go
export PATH=\$PATH:\$GOROOT/bin
export GOPATH=/home/$APP_USER/go
export PATH=\$PATH:\$GOPATH/bin
EOF

    chmod +x /etc/profile.d/go.sh
    source /etc/profile.d/go.sh

    # Verify installation
    /usr/local/go/bin/go version

    print_success "Go $GO_VERSION installed successfully"
}

# Install PostgreSQL
install_postgresql() {
    print_header "Installing PostgreSQL $POSTGRES_VERSION"

    if command -v psql &> /dev/null; then
        print_info "PostgreSQL is already installed"
        return
    fi

    if [ "$PKG_MANAGER" = "apt" ]; then
        # Ubuntu/Debian
        wget -qO - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
        echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list
        apt update
        $PKG_INSTALL postgresql-$POSTGRES_VERSION postgresql-contrib-$POSTGRES_VERSION
    else
        # CentOS/RHEL/Rocky/Fedora
        $PKG_INSTALL postgresql-server postgresql-contrib
        postgresql-setup --initdb
    fi

    # Start and enable PostgreSQL
    systemctl start postgresql
    systemctl enable postgresql

    print_success "PostgreSQL installed and started"
}

# Install Redis
install_redis() {
    print_header "Installing Redis"

    if command -v redis-server &> /dev/null; then
        print_info "Redis is already installed"
        return
    fi

    if [ "$PKG_MANAGER" = "apt" ]; then
        $PKG_INSTALL redis-server
    else
        $PKG_INSTALL redis
    fi

    # Start and enable Redis
    systemctl start redis
    systemctl enable redis

    print_success "Redis installed and started"
}

# Create application user
create_app_user() {
    print_header "Creating Application User"

    if id "$APP_USER" &>/dev/null; then
        print_info "User $APP_USER already exists"
    else
        useradd -r -s /bin/bash -d $INSTALL_DIR -m $APP_USER
        print_success "User $APP_USER created"
    fi
}

# Create directories
create_directories() {
    print_header "Creating Application Directories"

    mkdir -p $INSTALL_DIR
    mkdir -p $CONFIG_DIR
    mkdir -p $LOG_DIR
    mkdir -p $DATA_DIR
    mkdir -p $DATA_DIR/pcap
    mkdir -p $DATA_DIR/db

    # Set ownership
    chown -R $APP_USER:$APP_GROUP $INSTALL_DIR
    chown -R $APP_USER:$APP_GROUP $CONFIG_DIR
    chown -R $APP_USER:$APP_GROUP $LOG_DIR
    chown -R $APP_USER:$APP_GROUP $DATA_DIR

    # Set permissions
    chmod 750 $CONFIG_DIR
    chmod 755 $LOG_DIR
    chmod 755 $DATA_DIR

    print_success "Directories created"
}

# Setup PostgreSQL database
setup_database() {
    print_header "Setting up PostgreSQL Database"

    # Generate random password
    DB_PASSWORD=$(openssl rand -base64 32)

    # Create database and user
    sudo -u postgres psql <<EOF
-- Create user
CREATE USER protei_user WITH PASSWORD '$DB_PASSWORD';

-- Create database
CREATE DATABASE protei_monitoring OWNER protei_user;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE protei_monitoring TO protei_user;

-- Connect to database and create extensions
\c protei_monitoring
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

EOF

    # Save credentials
    cat > $CONFIG_DIR/db_credentials.txt <<EOF
Database: protei_monitoring
Username: protei_user
Password: $DB_PASSWORD
Host: localhost
Port: 5432
EOF

    chmod 600 $CONFIG_DIR/db_credentials.txt
    chown $APP_USER:$APP_GROUP $CONFIG_DIR/db_credentials.txt

    print_success "Database setup completed"
    print_warning "Database credentials saved to: $CONFIG_DIR/db_credentials.txt"
}

# Deploy application
deploy_application() {
    print_header "Deploying Application"

    # Check if encrypted package exists
    if [ ! -f "protei-monitoring-encrypted.tar.gz.enc" ]; then
        print_error "Encrypted package 'protei-monitoring-encrypted.tar.gz.enc' not found"
        print_info "Please ensure the encrypted package is in the current directory"
        exit 1
    fi

    print_info "Found encrypted package"
    print_warning "You will need the decryption key to proceed"

    # Prompt for decryption key
    read -sp "Enter decryption key: " DECRYPT_KEY
    echo

    # Decrypt package
    print_info "Decrypting package..."
    openssl enc -aes-256-cbc -d -in protei-monitoring-encrypted.tar.gz.enc -out /tmp/protei-monitoring.tar.gz -k "$DECRYPT_KEY"

    if [ $? -ne 0 ]; then
        print_error "Decryption failed. Invalid key or corrupted file"
        exit 1
    fi

    # Extract to installation directory
    print_info "Extracting application..."
    tar -xzf /tmp/protei-monitoring.tar.gz -C $INSTALL_DIR

    # Clean up
    rm -f /tmp/protei-monitoring.tar.gz

    # Set ownership
    chown -R $APP_USER:$APP_GROUP $INSTALL_DIR

    # Make binary executable
    chmod +x $INSTALL_DIR/bin/protei-monitoring

    print_success "Application deployed successfully"
}

# Create configuration file
create_config() {
    print_header "Creating Configuration File"

    # Read database password
    DB_PASSWORD=$(grep "Password:" $CONFIG_DIR/db_credentials.txt | awk '{print $2}')

    # Generate license key
    LICENSE_KEY=$(openssl rand -hex 32)

    cat > $CONFIG_DIR/config.yaml <<EOF
# Protei Monitoring Configuration
# Generated: $(date)

server:
  host: "0.0.0.0"
  port: 8443
  tls:
    enabled: true
    cert_file: "$CONFIG_DIR/certs/server.crt"
    key_file: "$CONFIG_DIR/certs/server.key"

database:
  host: "localhost"
  port: 5432
  database: "protei_monitoring"
  username: "protei_user"
  password: "$DB_PASSWORD"
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
  pcap_dir: "$DATA_DIR/pcap"
  buffer_size: 67108864  # 64MB
  max_file_size: 1073741824  # 1GB
  rotation_interval: 3600  # 1 hour

logging:
  level: "info"
  file: "$LOG_DIR/protei-monitoring.log"
  max_size: 100  # MB
  max_backups: 10
  max_age: 30  # days
  compress: true

license:
  key: "$LICENSE_KEY"
  company: "Your Company Name"
  max_users: 100
  features:
    - "protocol_decode"
    - "ai_analysis"
    - "flow_reconstruction"
    - "subscriber_correlation"

security:
  jwt_secret: "$(openssl rand -base64 64)"
  session_timeout: 3600  # 1 hour
  max_login_attempts: 5
  lockout_duration: 1800  # 30 minutes

monitoring:
  enable_metrics: true
  metrics_port: 9090
  health_check_interval: 30

EOF

    chmod 600 $CONFIG_DIR/config.yaml
    chown $APP_USER:$APP_GROUP $CONFIG_DIR/config.yaml

    print_success "Configuration file created: $CONFIG_DIR/config.yaml"
}

# Generate SSL certificates
generate_ssl_certs() {
    print_header "Generating SSL Certificates"

    mkdir -p $CONFIG_DIR/certs

    # Generate self-signed certificate
    openssl req -x509 -newkey rsa:4096 -nodes \
        -keyout $CONFIG_DIR/certs/server.key \
        -out $CONFIG_DIR/certs/server.crt \
        -days 365 \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=protei-monitoring"

    chmod 600 $CONFIG_DIR/certs/server.key
    chmod 644 $CONFIG_DIR/certs/server.crt
    chown -R $APP_USER:$APP_GROUP $CONFIG_DIR/certs

    print_success "SSL certificates generated"
    print_warning "Using self-signed certificate. Replace with valid certificate for production"
}

# Create systemd service
create_systemd_service() {
    print_header "Creating Systemd Service"

    cat > /etc/systemd/system/protei-monitoring.service <<EOF
[Unit]
Description=Protei Monitoring - Telecom Protocol Monitoring System
Documentation=https://github.com/protei/monitoring
After=network.target postgresql.service redis.service
Requires=postgresql.service redis.service

[Service]
Type=simple
User=$APP_USER
Group=$APP_GROUP
WorkingDirectory=$INSTALL_DIR

# Environment
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
Environment="CONFIG_FILE=$CONFIG_DIR/config.yaml"

# Executable
ExecStart=$INSTALL_DIR/bin/protei-monitoring -config $CONFIG_DIR/config.yaml

# Security
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Restart policy
Restart=always
RestartSec=10
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=protei-monitoring

# Limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    systemctl daemon-reload

    print_success "Systemd service created"
}

# Configure firewall
configure_firewall() {
    print_header "Configuring Firewall"

    # Check if firewalld is running
    if systemctl is-active --quiet firewalld; then
        print_info "Configuring firewalld..."
        firewall-cmd --permanent --add-port=8443/tcp
        firewall-cmd --permanent --add-port=9090/tcp
        firewall-cmd --reload
        print_success "Firewalld configured"
    # Check if ufw is running
    elif systemctl is-active --quiet ufw; then
        print_info "Configuring ufw..."
        ufw allow 8443/tcp
        ufw allow 9090/tcp
        print_success "UFW configured"
    else
        print_warning "No firewall detected. Please configure firewall manually:"
        print_info "  - Allow TCP port 8443 (HTTPS Web Interface)"
        print_info "  - Allow TCP port 9090 (Metrics)"
    fi
}

# Set permissions for packet capture
set_capture_permissions() {
    print_header "Setting Packet Capture Permissions"

    # Give capability to capture packets
    setcap 'cap_net_raw,cap_net_admin+eip' $INSTALL_DIR/bin/protei-monitoring

    print_success "Packet capture permissions set"
}

# Create admin user
create_admin_user() {
    print_header "Creating Admin User"

    print_info "Default admin credentials will be created"
    print_warning "Please change the password after first login"

    # Generate random password
    ADMIN_PASSWORD=$(openssl rand -base64 16)

    cat > $CONFIG_DIR/admin_credentials.txt <<EOF
Username: admin
Password: $ADMIN_PASSWORD
URL: https://$(hostname -I | awk '{print $1}'):8443

IMPORTANT: Change this password after first login!
EOF

    chmod 600 $CONFIG_DIR/admin_credentials.txt
    chown $APP_USER:$APP_GROUP $CONFIG_DIR/admin_credentials.txt

    print_success "Admin credentials created"
    print_warning "Credentials saved to: $CONFIG_DIR/admin_credentials.txt"
}

# Start service
start_service() {
    print_header "Starting Protei Monitoring Service"

    systemctl enable protei-monitoring
    systemctl start protei-monitoring

    # Wait for service to start
    sleep 3

    if systemctl is-active --quiet protei-monitoring; then
        print_success "Service started successfully"
    else
        print_error "Service failed to start"
        print_info "Check logs: journalctl -u protei-monitoring -n 50"
        exit 1
    fi
}

# Print summary
print_summary() {
    print_header "Installation Complete!"

    echo -e "${GREEN}Protei Monitoring has been successfully installed!${NC}\n"

    echo -e "${YELLOW}Important Files:${NC}"
    echo -e "  Configuration: $CONFIG_DIR/config.yaml"
    echo -e "  Admin Credentials: $CONFIG_DIR/admin_credentials.txt"
    echo -e "  Database Credentials: $CONFIG_DIR/db_credentials.txt"
    echo -e "  Logs: $LOG_DIR/protei-monitoring.log"
    echo -e "  Data: $DATA_DIR"
    echo ""

    echo -e "${YELLOW}Service Management:${NC}"
    echo -e "  Start:   systemctl start protei-monitoring"
    echo -e "  Stop:    systemctl stop protei-monitoring"
    echo -e "  Restart: systemctl restart protei-monitoring"
    echo -e "  Status:  systemctl status protei-monitoring"
    echo -e "  Logs:    journalctl -u protei-monitoring -f"
    echo ""

    # Get IP address
    IP_ADDR=$(hostname -I | awk '{print $1}')

    echo -e "${YELLOW}Access Information:${NC}"
    echo -e "  Web Interface: https://$IP_ADDR:8443"
    echo -e "  Metrics:       http://$IP_ADDR:9090/metrics"
    echo ""

    # Show admin credentials
    echo -e "${YELLOW}Admin Credentials:${NC}"
    cat $CONFIG_DIR/admin_credentials.txt
    echo ""

    echo -e "${RED}SECURITY REMINDERS:${NC}"
    echo -e "  1. Change the admin password immediately"
    echo -e "  2. Replace self-signed SSL certificate with valid certificate"
    echo -e "  3. Secure the credentials files ($CONFIG_DIR/*.txt)"
    echo -e "  4. Review and adjust firewall rules"
    echo -e "  5. Configure backup for $DATA_DIR"
    echo ""

    print_success "Installation completed successfully!"
}

# Main installation flow
main() {
    print_header "Protei Monitoring - Automated Installation"

    check_root
    detect_os
    update_system
    install_dependencies
    install_go
    install_postgresql
    install_redis
    create_app_user
    create_directories
    setup_database
    deploy_application
    create_config
    generate_ssl_certs
    create_systemd_service
    configure_firewall
    set_capture_permissions
    create_admin_user
    start_service
    print_summary
}

# Run main function
main "$@"
