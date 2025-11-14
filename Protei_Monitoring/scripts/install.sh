#!/bin/bash
#
# Protei Monitoring v2.0 - Installation Script
#
# This script performs a complete installation of Protei Monitoring including:
# - Prerequisites verification
# - Database schema creation
# - User creation
# - Configuration validation
# - Initial data loading
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Installation directory
INSTALL_DIR="/usr/protei/Protei_Monitoring"
CONFIG_DIR="$INSTALL_DIR/config"
BIN_DIR="$INSTALL_DIR/bin"
LOG_DIR="$INSTALL_DIR/logs"
CDR_DIR="$INSTALL_DIR/cdr"
TMP_DIR="$INSTALL_DIR/tmp"

# Log file
INSTALL_LOG="$LOG_DIR/installation.log"

# Print functions
print_header() {
    echo -e "${BLUE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Protei Monitoring v2.0 - Installation"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] STEP: $1" >> "$INSTALL_LOG"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1" >> "$INSTALL_LOG"
}

print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >> "$INSTALL_LOG"
}

print_warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $1" >> "$INSTALL_LOG"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."

    local missing_deps=()

    # Check PostgreSQL
    if ! command -v psql &> /dev/null; then
        missing_deps+=("PostgreSQL client (psql)")
    fi

    # Check Redis
    if ! command -v redis-cli &> /dev/null; then
        missing_deps+=("Redis client (redis-cli)")
    fi

    # Check libpcap
    if ! ldconfig -p | grep -q libpcap; then
        missing_deps+=("libpcap library")
    fi

    # Check Go (if building from source)
    if ! command -v go &> /dev/null; then
        print_warning "Go compiler not found (required only for building from source)"
    fi

    if [ ${#missing_deps[@]} -gt 0 ]; then
        print_error "Missing dependencies:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        echo ""
        echo "Install missing dependencies:"
        echo "  RHEL/CentOS: sudo yum install postgresql-client redis libpcap"
        echo "  Ubuntu/Debian: sudo apt install postgresql-client redis-tools libpcap0.8"
        exit 1
    fi

    print_success "All prerequisites satisfied"
}

# Create directory structure
create_directories() {
    print_step "Creating directory structure..."

    # Create all required directories
    mkdir -p "$LOG_DIR"/{application,system,debug,error,access}
    mkdir -p "$CDR_DIR"/{MAP,CAP,INAP,Diameter,GTP,PFCP,HTTP2,NGAP,S1AP,NAS,combined}
    mkdir -p "$TMP_DIR"

    # Set permissions
    chmod 755 "$INSTALL_DIR"
    chmod 750 "$CONFIG_DIR"
    chmod 755 "$LOG_DIR"
    chmod 755 "$CDR_DIR"
    chmod 755 "$TMP_DIR"

    print_success "Directory structure created"
}

# Load configuration
load_config() {
    print_step "Loading configuration..."

    if [ ! -f "$CONFIG_DIR/db.cfg" ]; then
        print_error "Database configuration file not found: $CONFIG_DIR/db.cfg"
        exit 1
    fi

    # Source database configuration
    source "$CONFIG_DIR/db.cfg"

    # Validate configuration
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
        print_error "Database configuration incomplete. Please check $CONFIG_DIR/db.cfg"
        exit 1
    fi

    print_success "Configuration loaded"
}

# Test database connection
test_database() {
    print_step "Testing database connection..."

    export PGPASSWORD="$DB_PASSWORD"

    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT 1;" &> /dev/null; then
        print_error "Cannot connect to database. Please check:"
        echo "  - PostgreSQL is running"
        echo "  - Database credentials in $CONFIG_DIR/db.cfg"
        echo "  - Network connectivity to $DB_HOST:$DB_PORT"
        exit 1
    fi

    print_success "Database connection successful"
}

# Create database
create_database() {
    print_step "Creating database '$DB_NAME'..."

    export PGPASSWORD="$DB_PASSWORD"

    # Check if database exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        print_warning "Database '$DB_NAME' already exists. Skipping creation."
        return 0
    fi

    # Create database
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>> "$INSTALL_LOG"; then
        print_error "Failed to create database. Check if you have sufficient privileges."
        exit 1
    fi

    print_success "Database created"
}

# Create database schema
create_schema() {
    print_step "Creating database schema..."

    export PGPASSWORD="$DB_PASSWORD"

    # SQL schema
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<'EOF' >> "$INSTALL_LOG" 2>&1

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    email VARCHAR(100),
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'operator', 'viewer')),
    enabled BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id BIGSERIAL PRIMARY KEY,
    transaction_id VARCHAR(100) UNIQUE NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_ms INTEGER,
    imsi VARCHAR(15),
    msisdn VARCHAR(15),
    imei VARCHAR(15),
    source_ip INET,
    dest_ip INET,
    source_port INTEGER,
    dest_port INTEGER,
    message_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'ongoing',
    error_code INTEGER,
    error_description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES sessions(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    direction VARCHAR(10) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    source VARCHAR(100),
    destination VARCHAR(100),
    decoded_data JSONB,
    raw_data BYTEA,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Subscribers table
CREATE TABLE IF NOT EXISTS subscribers (
    id BIGSERIAL PRIMARY KEY,
    imsi VARCHAR(15) UNIQUE,
    msisdn VARCHAR(15),
    imei VARCHAR(15),
    last_location VARCHAR(100),
    last_seen TIMESTAMP,
    first_seen TIMESTAMP DEFAULT NOW(),
    total_sessions INTEGER DEFAULT 0,
    metadata JSONB
);

-- Issues table
CREATE TABLE IF NOT EXISTS issues (
    id BIGSERIAL PRIMARY KEY,
    severity VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    protocol VARCHAR(20),
    description TEXT NOT NULL,
    root_cause TEXT,
    recommendation TEXT,
    standard_ref VARCHAR(100),
    affected_sessions INTEGER DEFAULT 0,
    first_occurrence TIMESTAMP DEFAULT NOW(),
    last_occurrence TIMESTAMP DEFAULT NOW(),
    count INTEGER DEFAULT 1,
    resolved BOOLEAN DEFAULT false
);

-- KPIs table
CREATE TABLE IF NOT EXISTS kpis (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    protocol VARCHAR(20),
    metric_name VARCHAR(50) NOT NULL,
    metric_value NUMERIC(12,2),
    metadata JSONB
);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    details JSONB,
    ip_address INET,
    timestamp TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sessions_imsi ON sessions(imsi);
CREATE INDEX IF NOT EXISTS idx_sessions_msisdn ON sessions(msisdn);
CREATE INDEX IF NOT EXISTS idx_sessions_protocol ON sessions(protocol);
CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_subscribers_imsi ON subscribers(imsi);
CREATE INDEX IF NOT EXISTS idx_subscribers_msisdn ON subscribers(msisdn);
CREATE INDEX IF NOT EXISTS idx_issues_severity ON issues(severity);
CREATE INDEX IF NOT EXISTS idx_kpis_timestamp ON kpis(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_kpis_protocol ON kpis(protocol);

EOF

    if [ $? -eq 0 ]; then
        print_success "Database schema created"
    else
        print_error "Failed to create database schema. Check $INSTALL_LOG for details."
        exit 1
    fi
}

# Create default admin user
create_admin_user() {
    print_step "Creating default admin user..."

    export PGPASSWORD="$DB_PASSWORD"

    # Check if admin user exists
    ADMIN_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE username='admin';")

    if [ "$ADMIN_EXISTS" -gt 0 ]; then
        print_warning "Admin user already exists. Skipping creation."
        return 0
    fi

    # Default password: "admin" (should be changed after first login)
    # BCrypt hash of "admin"
    ADMIN_HASH='$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
        "INSERT INTO users (username, password_hash, full_name, email, role) VALUES ('admin', '$ADMIN_HASH', 'System Administrator', 'admin@protei.com', 'admin');" \
        >> "$INSTALL_LOG" 2>&1

    if [ $? -eq 0 ]; then
        print_success "Admin user created (username: admin, password: admin)"
        print_warning "IMPORTANT: Change the default admin password after first login!"
    else
        print_error "Failed to create admin user"
        exit 1
    fi
}

# Validate configuration files
validate_config() {
    print_step "Validating configuration files..."

    local config_files=("license.cfg" "db.cfg" "protocols.cfg" "system.cfg" "trace.cfg" "paths.cfg" "security.cfg")
    local missing_files=()

    for cfg in "${config_files[@]}"; do
        if [ ! -f "$CONFIG_DIR/$cfg" ]; then
            missing_files+=("$cfg")
        else
            # Validate bash syntax
            if ! bash -n "$CONFIG_DIR/$cfg" 2>> "$INSTALL_LOG"; then
                print_error "Syntax error in $cfg"
                exit 1
            fi
        fi
    done

    if [ ${#missing_files[@]} -gt 0 ]; then
        print_error "Missing configuration files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        exit 1
    fi

    print_success "All configuration files validated"
}

# Set file permissions
set_permissions() {
    print_step "Setting file permissions..."

    # Create protei user and group if they don't exist
    if ! id -u protei &>/dev/null; then
        useradd -r -s /bin/false protei
        print_success "Created protei user"
    fi

    # Set ownership
    chown -R root:protei "$INSTALL_DIR"
    chown -R protei:protei "$LOG_DIR"
    chown -R protei:protei "$CDR_DIR"
    chown -R protei:protei "$TMP_DIR"

    # Set permissions
    chmod -R 750 "$CONFIG_DIR"
    chmod -R 755 "$BIN_DIR"
    chmod -R 755 "$LOG_DIR"
    chmod -R 755 "$CDR_DIR"
    chmod -R 755 "$TMP_DIR"

    # Protect sensitive configs
    chmod 600 "$CONFIG_DIR/db.cfg"
    chmod 600 "$CONFIG_DIR/license.cfg"
    chmod 600 "$CONFIG_DIR/security.cfg"

    print_success "File permissions set"
}

# Create systemd service (optional)
create_systemd_service() {
    print_step "Creating systemd service..."

    cat > /etc/systemd/system/protei-monitoring.service <<EOF
[Unit]
Description=Protei Monitoring v2.0
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=protei
Group=protei
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/scripts/start
ExecStop=$INSTALL_DIR/scripts/stop
ExecReload=$INSTALL_DIR/scripts/reload
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    print_success "Systemd service created"
    echo "  To enable auto-start: systemctl enable protei-monitoring"
    echo "  To start now: systemctl start protei-monitoring"
}

# Test Redis connection
test_redis() {
    print_step "Testing Redis connection..."

    source "$CONFIG_DIR/system.cfg"

    if [ "$REDIS_ENABLED" != "true" ]; then
        print_warning "Redis is disabled in configuration"
        return 0
    fi

    if ! redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &> /dev/null; then
        print_warning "Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
        echo "  Redis is optional but recommended for caching"
        return 0
    fi

    print_success "Redis connection successful"
}

# Print installation summary
print_summary() {
    echo ""
    echo -e "${GREEN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Installation Complete!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
    echo ""
    echo "Installation Directory: $INSTALL_DIR"
    echo "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
    echo "Default Admin: admin / admin"
    echo ""
    echo "Next Steps:"
    echo "  1. Update license configuration: sudo nano $CONFIG_DIR/license.cfg"
    echo "  2. Change admin password (via web UI after first login)"
    echo "  3. Review configuration files in: $CONFIG_DIR/"
    echo "  4. Start application: sudo $INSTALL_DIR/scripts/start"
    echo "  5. Access web interface: http://<server_ip>:8080"
    echo ""
    echo "Documentation: $INSTALL_DIR/document/"
    echo "Logs: $LOG_DIR/"
    echo ""
    print_success "Protei Monitoring v2.0 is ready to use!"
}

# Main installation flow
main() {
    print_header

    # Ensure log directory exists
    mkdir -p "$LOG_DIR"
    echo "Installation started at $(date)" > "$INSTALL_LOG"

    check_root
    check_prerequisites
    create_directories
    load_config
    test_database
    create_database
    create_schema
    create_admin_user
    test_redis
    validate_config
    set_permissions
    create_systemd_service

    print_summary
}

# Run installation
main

exit 0
