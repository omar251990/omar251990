#!/bin/bash
################################################################################
# Protei_Bulk - Automated Installation Script
# Version: 1.0.0
#
# This script automatically installs and configures Protei_Bulk platform:
# - System dependencies
# - PostgreSQL database
# - Database user and schema
# - Python environment
# - Application configuration
# - Service setup
################################################################################

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Installation settings
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_USER="protei"
DB_PASSWORD="elephant"
DB_NAME="protei_bulk"
DB_HOST="localhost"
DB_PORT="5432"
REDIS_PORT="6379"
APP_USER="protei"
PYTHON_VERSION="3.8"

# Log file
LOG_FILE="$INSTALL_DIR/installation.log"

################################################################################
# Logging Functions
################################################################################

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${CYAN}[i]${NC} $1" | tee -a "$LOG_FILE"
}

################################################################################
# Helper Functions
################################################################################

print_header() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  $1"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

check_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
        log_info "Detected OS: $OS $OS_VERSION"
    else
        log_error "Cannot detect operating system"
        exit 1
    fi

    # Support Ubuntu 18.04+, Debian 10+, CentOS 7+
    case $OS in
        ubuntu|debian)
            PKG_MANAGER="apt-get"
            ;;
        centos|rhel)
            PKG_MANAGER="yum"
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

confirm_installation() {
    echo ""
    print_header "Protei_Bulk Installation"
    echo "This script will install and configure Protei_Bulk with the following:"
    echo ""
    echo "  • PostgreSQL database server"
    echo "  • Redis server"
    echo "  • Python $PYTHON_VERSION and dependencies"
    echo "  • Database: $DB_NAME"
    echo "  • Database User: $DB_USER"
    echo "  • Installation Directory: $INSTALL_DIR"
    echo ""
    read -p "Do you want to continue? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warning "Installation cancelled by user"
        exit 0
    fi
}

################################################################################
# System Dependencies Installation
################################################################################

install_system_dependencies() {
    print_header "Installing System Dependencies"

    log "Updating package list..."
    if [ "$PKG_MANAGER" = "apt-get" ]; then
        apt-get update -qq
    else
        yum update -y -q
    fi

    log "Installing required packages..."

    if [ "$PKG_MANAGER" = "apt-get" ]; then
        DEBIAN_FRONTEND=noninteractive apt-get install -y \
            build-essential \
            python3 \
            python3-pip \
            python3-dev \
            python3-venv \
            postgresql \
            postgresql-contrib \
            postgresql-server-dev-all \
            redis-server \
            libpq-dev \
            libssl-dev \
            libffi-dev \
            git \
            curl \
            wget \
            supervisor \
            nginx \
            net-tools \
            htop \
            vim \
            >> "$LOG_FILE" 2>&1
    else
        yum install -y \
            gcc \
            gcc-c++ \
            make \
            python3 \
            python3-pip \
            python3-devel \
            postgresql-server \
            postgresql-contrib \
            postgresql-devel \
            redis \
            openssl-devel \
            libffi-devel \
            git \
            curl \
            wget \
            supervisor \
            nginx \
            net-tools \
            htop \
            vim \
            >> "$LOG_FILE" 2>&1
    fi

    log_success "System dependencies installed"
}

################################################################################
# PostgreSQL Setup
################################################################################

setup_postgresql() {
    print_header "Setting Up PostgreSQL"

    # Initialize PostgreSQL if needed (CentOS/RHEL)
    if [ "$PKG_MANAGER" = "yum" ]; then
        if [ ! -d "/var/lib/pgsql/data" ]; then
            log "Initializing PostgreSQL database..."
            postgresql-setup initdb >> "$LOG_FILE" 2>&1
        fi
    fi

    # Start PostgreSQL
    log "Starting PostgreSQL service..."
    if [ "$PKG_MANAGER" = "apt-get" ]; then
        systemctl start postgresql
        systemctl enable postgresql
    else
        systemctl start postgresql
        systemctl enable postgresql
    fi

    # Wait for PostgreSQL to start
    sleep 3

    log_success "PostgreSQL service started"

    # Create database user
    log "Creating database user: $DB_USER..."
    sudo -u postgres psql -c "DROP USER IF EXISTS $DB_USER;" >> "$LOG_FILE" 2>&1 || true
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" >> "$LOG_FILE" 2>&1
    sudo -u postgres psql -c "ALTER USER $DB_USER WITH SUPERUSER;" >> "$LOG_FILE" 2>&1

    log_success "Database user created: $DB_USER"

    # Create database
    log "Creating database: $DB_NAME..."
    sudo -u postgres psql -c "DROP DATABASE IF EXISTS $DB_NAME;" >> "$LOG_FILE" 2>&1 || true
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" >> "$LOG_FILE" 2>&1

    log_success "Database created: $DB_NAME"

    # Configure PostgreSQL for local connections
    log "Configuring PostgreSQL authentication..."

    if [ "$PKG_MANAGER" = "apt-get" ]; then
        PG_HBA="/etc/postgresql/$(ls /etc/postgresql | head -1)/main/pg_hba.conf"
        PG_CONF="/etc/postgresql/$(ls /etc/postgresql | head -1)/main/postgresql.conf"
    else
        PG_HBA="/var/lib/pgsql/data/pg_hba.conf"
        PG_CONF="/var/lib/pgsql/data/postgresql.conf"
    fi

    # Backup original pg_hba.conf
    cp "$PG_HBA" "$PG_HBA.backup" 2>/dev/null || true

    # Add authentication rule for protei user
    if ! grep -q "host.*$DB_NAME.*$DB_USER" "$PG_HBA"; then
        echo "# Protei_Bulk database access" >> "$PG_HBA"
        echo "host    $DB_NAME    $DB_USER    127.0.0.1/32    md5" >> "$PG_HBA"
        echo "host    $DB_NAME    $DB_USER    ::1/128         md5" >> "$PG_HBA"
    fi

    # Restart PostgreSQL to apply changes
    log "Restarting PostgreSQL..."
    systemctl restart postgresql
    sleep 3

    log_success "PostgreSQL configured"
}

################################################################################
# Database Schema Setup
################################################################################

setup_database_schema() {
    print_header "Setting Up Database Schema"

    SCHEMA_FILE="$INSTALL_DIR/database/schema.sql"

    if [ ! -f "$SCHEMA_FILE" ]; then
        log_error "Schema file not found: $SCHEMA_FILE"
        exit 1
    fi

    log "Loading database schema..."

    # Set PostgreSQL password for connection
    export PGPASSWORD="$DB_PASSWORD"

    # Load schema
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SCHEMA_FILE" >> "$LOG_FILE" 2>&1

    if [ $? -eq 0 ]; then
        log_success "Database schema loaded successfully"
    else
        log_error "Failed to load database schema"
        exit 1
    fi

    # Count tables created
    TABLE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
    log_info "Created $TABLE_COUNT database tables"

    unset PGPASSWORD
}

################################################################################
# Redis Setup
################################################################################

setup_redis() {
    print_header "Setting Up Redis"

    log "Starting Redis service..."
    systemctl start redis-server 2>/dev/null || systemctl start redis 2>/dev/null
    systemctl enable redis-server 2>/dev/null || systemctl enable redis 2>/dev/null

    # Test Redis connection
    if redis-cli ping > /dev/null 2>&1; then
        log_success "Redis service started and responding"
    else
        log_warning "Redis may not be running properly"
    fi
}

################################################################################
# Python Environment Setup
################################################################################

setup_python_environment() {
    print_header "Setting Up Python Environment"

    # Upgrade pip
    log "Upgrading pip..."
    python3 -m pip install --upgrade pip >> "$LOG_FILE" 2>&1

    # Create virtual environment
    log "Creating Python virtual environment..."
    VENV_DIR="$INSTALL_DIR/venv"

    if [ -d "$VENV_DIR" ]; then
        log_warning "Virtual environment already exists, removing..."
        rm -rf "$VENV_DIR"
    fi

    python3 -m venv "$VENV_DIR" >> "$LOG_FILE" 2>&1

    # Activate virtual environment
    source "$VENV_DIR/bin/activate"

    # Install Python dependencies
    log "Installing Python dependencies..."
    REQUIREMENTS_FILE="$INSTALL_DIR/requirements.txt"

    if [ -f "$REQUIREMENTS_FILE" ]; then
        pip install -r "$REQUIREMENTS_FILE" >> "$LOG_FILE" 2>&1
        log_success "Python dependencies installed"
    else
        log_warning "requirements.txt not found, skipping Python dependencies"
    fi

    deactivate
}

################################################################################
# Configuration Setup
################################################################################

setup_configuration() {
    print_header "Setting Up Configuration"

    CONFIG_DIR="$INSTALL_DIR/config"

    # Update database configuration
    log "Configuring database connection..."

    DB_CONF="$CONFIG_DIR/db.conf"

    if [ -f "$DB_CONF" ]; then
        # Update PostgreSQL settings
        sed -i "s/^host = .*/host = $DB_HOST/" "$DB_CONF"
        sed -i "s/^port = .*/port = $DB_PORT/" "$DB_CONF"
        sed -i "s/^database = .*/database = $DB_NAME/" "$DB_CONF"
        sed -i "s/^username = .*/username = $DB_USER/" "$DB_CONF"
        sed -i "s/^password = .*/password = $DB_PASSWORD/" "$DB_CONF"

        log_success "Database configuration updated"
    fi

    # Update application configuration
    APP_CONF="$CONFIG_DIR/app.conf"

    if [ -f "$APP_CONF" ]; then
        sed -i "s|^base_dir = .*|base_dir = $INSTALL_DIR|" "$APP_CONF"
        log_success "Application configuration updated"
    fi

    # Set proper permissions
    log "Setting file permissions..."
    chmod 600 "$CONFIG_DIR"/*.conf
    chmod 600 "$CONFIG_DIR/license.key"

    # Create necessary directories
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "$INSTALL_DIR/tmp"
    mkdir -p "$INSTALL_DIR/cdr/smpp"
    mkdir -p "$INSTALL_DIR/cdr/http"
    mkdir -p "$INSTALL_DIR/cdr/internal"
    mkdir -p "$INSTALL_DIR/cdr/archive"

    log_success "Configuration completed"
}

################################################################################
# Application User Setup
################################################################################

setup_application_user() {
    print_header "Setting Up Application User"

    # Create application user if doesn't exist
    if ! id "$APP_USER" &>/dev/null; then
        log "Creating application user: $APP_USER..."
        useradd -r -s /bin/bash -d "$INSTALL_DIR" "$APP_USER" 2>/dev/null || true
        log_success "User created: $APP_USER"
    else
        log_info "User already exists: $APP_USER"
    fi

    # Set ownership
    log "Setting file ownership..."
    chown -R "$APP_USER:$APP_USER" "$INSTALL_DIR"

    log_success "Ownership configured"
}

################################################################################
# System Service Setup
################################################################################

setup_systemd_service() {
    print_header "Setting Up System Service"

    SERVICE_FILE="/etc/systemd/system/protei_bulk.service"

    log "Creating systemd service..."

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Protei Bulk Messaging Platform
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=forking
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$INSTALL_DIR
Environment="PATH=$INSTALL_DIR/venv/bin:/usr/local/bin:/usr/bin:/bin"
ExecStart=$INSTALL_DIR/scripts/start
ExecStop=$INSTALL_DIR/scripts/stop
ExecReload=$INSTALL_DIR/scripts/reload
PIDFile=$INSTALL_DIR/tmp/protei_bulk.pid
Restart=on-failure
RestartSec=10
StandardOutput=append:$INSTALL_DIR/logs/system.log
StandardError=append:$INSTALL_DIR/logs/error.log

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    systemctl daemon-reload

    log_success "Systemd service created"
}

################################################################################
# Post-Installation Verification
################################################################################

verify_installation() {
    print_header "Verifying Installation"

    ERRORS=0

    # Check PostgreSQL
    log "Checking PostgreSQL..."
    if systemctl is-active --quiet postgresql; then
        log_success "PostgreSQL is running"
    else
        log_error "PostgreSQL is not running"
        ((ERRORS++))
    fi

    # Check Redis
    log "Checking Redis..."
    if systemctl is-active --quiet redis-server || systemctl is-active --quiet redis; then
        log_success "Redis is running"
    else
        log_error "Redis is not running"
        ((ERRORS++))
    fi

    # Check database connection
    log "Checking database connection..."
    export PGPASSWORD="$DB_PASSWORD"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "Database connection successful"
    else
        log_error "Cannot connect to database"
        ((ERRORS++))
    fi
    unset PGPASSWORD

    # Check Python environment
    log "Checking Python environment..."
    if [ -f "$INSTALL_DIR/venv/bin/python" ]; then
        PYTHON_VER=$("$INSTALL_DIR/venv/bin/python" --version 2>&1)
        log_success "Python environment ready: $PYTHON_VER"
    else
        log_error "Python virtual environment not found"
        ((ERRORS++))
    fi

    # Check configuration files
    log "Checking configuration files..."
    CONFIG_FILES=("app.conf" "db.conf" "log.conf" "protocol.conf" "network.conf" "security.conf")
    for conf in "${CONFIG_FILES[@]}"; do
        if [ -f "$CONFIG_DIR/$conf" ]; then
            log_success "$conf exists"
        else
            log_error "$conf missing"
            ((ERRORS++))
        fi
    done

    # Check directories
    log "Checking directories..."
    REQUIRED_DIRS=("bin" "config" "database" "logs" "tmp" "cdr" "scripts")
    for dir in "${REQUIRED_DIRS[@]}"; do
        if [ -d "$INSTALL_DIR/$dir" ]; then
            log_success "$dir/ exists"
        else
            log_error "$dir/ missing"
            ((ERRORS++))
        fi
    done

    echo ""
    if [ $ERRORS -eq 0 ]; then
        log_success "All verification checks passed!"
        return 0
    else
        log_error "Verification found $ERRORS error(s)"
        return 1
    fi
}

################################################################################
# Installation Summary
################################################################################

print_installation_summary() {
    print_header "Installation Summary"

    echo ""
    echo "  Installation Directory: $INSTALL_DIR"
    echo "  Database Name: $DB_NAME"
    echo "  Database User: $DB_USER"
    echo "  Database Host: $DB_HOST:$DB_PORT"
    echo "  Application User: $APP_USER"
    echo "  Log File: $LOG_FILE"
    echo ""
    echo "  Configuration Files:"
    echo "    • $CONFIG_DIR/app.conf"
    echo "    • $CONFIG_DIR/db.conf"
    echo "    • $CONFIG_DIR/protocol.conf"
    echo "    • $CONFIG_DIR/network.conf"
    echo "    • $CONFIG_DIR/security.conf"
    echo ""
    echo "  Management Commands:"
    echo "    • Start:   sudo systemctl start protei_bulk"
    echo "    • Stop:    sudo systemctl stop protei_bulk"
    echo "    • Status:  sudo systemctl status protei_bulk"
    echo "    • Logs:    tail -f $INSTALL_DIR/logs/system.log"
    echo ""
    echo "  Or use the scripts directly:"
    echo "    • $INSTALL_DIR/scripts/start"
    echo "    • $INSTALL_DIR/scripts/stop"
    echo "    • $INSTALL_DIR/scripts/status"
    echo ""
}

################################################################################
# Main Installation Flow
################################################################################

main() {
    clear

    # Start logging
    echo "Protei_Bulk Installation Started at $(date)" > "$LOG_FILE"

    # Pre-installation checks
    check_root
    check_os
    confirm_installation

    # Installation steps
    install_system_dependencies
    setup_postgresql
    setup_database_schema
    setup_redis
    setup_python_environment
    setup_configuration
    setup_application_user
    setup_systemd_service

    # Verification
    verify_installation
    VERIFY_RESULT=$?

    # Summary
    print_installation_summary

    if [ $VERIFY_RESULT -eq 0 ]; then
        echo -e "${GREEN}"
        echo "╔════════════════════════════════════════════════════════════════╗"
        echo "║                                                                ║"
        echo "║          ✓ Installation Completed Successfully!               ║"
        echo "║                                                                ║"
        echo "╚════════════════════════════════════════════════════════════════╝"
        echo -e "${NC}"
        echo ""
        echo "Next steps:"
        echo "  1. Review configuration files in $CONFIG_DIR"
        echo "  2. Update license.key with your valid license"
        echo "  3. Start the service: sudo systemctl start protei_bulk"
        echo "  4. Check status: sudo systemctl status protei_bulk"
        echo ""
        exit 0
    else
        echo -e "${RED}"
        echo "╔════════════════════════════════════════════════════════════════╗"
        echo "║                                                                ║"
        echo "║          ✗ Installation Completed with Errors                 ║"
        echo "║                                                                ║"
        echo "╚════════════════════════════════════════════════════════════════╝"
        echo -e "${NC}"
        echo ""
        echo "Please check the log file for details: $LOG_FILE"
        echo ""
        exit 1
    fi
}

# Run main installation
main "$@"
