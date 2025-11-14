# config/ - Configuration Files

All application configuration files are stored in this directory.

## Configuration Files

### 1. license.cfg - License Configuration

**Purpose**: License validation and feature enablement

**Key Parameters:**
```bash
LICENSE_TYPE="enterprise"              # License type
LICENSE_KEY="XXXX-XXXX-XXXX-XXXX"     # License key
LICENSE_EXPIRY="2026-12-31"            # Expiry date (YYYY-MM-DD)
LICENSE_MAC="XX:XX:XX:XX:XX:XX"        # Server MAC address binding
LICENSE_PROTOCOLS="MAP,CAP,INAP,Diameter,GTP,PFCP,HTTP2,NGAP,S1AP,NAS"
LICENSE_MAX_SESSIONS=10000             # Maximum concurrent sessions
LICENSE_MAX_USERS=50                   # Maximum system users
```

**How to Update:**
```bash
# Get your server's MAC address
ip link show | grep ether

# Edit license file
sudo nano license.cfg

# Update LICENSE_MAC with your server's MAC
LICENSE_MAC="aa:bb:cc:dd:ee:ff"

# Reload configuration
sudo ../scripts/reload
```

---

### 2. db.cfg - Database Configuration

**Purpose**: PostgreSQL database connection settings

**Key Parameters:**
```bash
DB_ENABLED=true
DB_HOST="localhost"                    # Database server hostname
DB_PORT=5432                           # Database port
DB_NAME="protei_monitoring"            # Database name
DB_USER="protei"                       # Database username
DB_PASSWORD="<secure_password>"        # Database password
DB_POOL_SIZE=50                        # Connection pool size
DB_MAX_IDLE=10                         # Maximum idle connections
DB_CONN_LIFETIME=3600                  # Connection lifetime (seconds)
DB_SSL_MODE="disable"                  # SSL mode (disable/require/verify-full)
```

**Database Setup:**
```bash
# Create database
sudo -u postgres psql <<EOF
CREATE DATABASE protei_monitoring;
CREATE USER protei WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE protei_monitoring TO protei;
EOF

# Update db.cfg
sudo nano db.cfg
# Set DB_PASSWORD to your password

# Test connection
psql -h localhost -U protei -d protei_monitoring -c "SELECT 1;"
```

---

### 3. protocols.cfg - Protocol Configuration

**Purpose**: Enable/disable protocols and configure protocol-specific settings

**Sections:**

#### Protocol Enable/Disable
```bash
MAP_ENABLED=true
CAP_ENABLED=true
INAP_ENABLED=true
DIAMETER_ENABLED=true
GTP_ENABLED=true
PFCP_ENABLED=true
HTTP2_ENABLED=true
NGAP_ENABLED=true
S1AP_ENABLED=true
NAS_ENABLED=true
```

#### Protocol-Specific Settings

**MAP Protocol:**
```bash
MAP_VERSION=3                          # MAP version (2 or 3)
MAP_DECODE_FULL=true                   # Full decode vs. header only
MAP_SAVE_CDR=true                      # Save to CDR files
MAP_CDR_PATH="cdr/MAP/"               # CDR output directory
```

**Diameter Protocol:**
```bash
DIAMETER_APPLICATIONS="S6a,S6d,SWm,SWx,Gx,Gy"  # Supported applications
DIAMETER_DECODE_AVP=true               # Decode AVPs
DIAMETER_SAVE_CDR=true
DIAMETER_CDR_PATH="cdr/Diameter/"
```

**GTP Protocol:**
```bash
GTP_VERSION=2                          # GTPv1 or GTPv2
GTP_TRACK_TUNNELS=true                 # Track tunnel states
GTP_SAVE_CDR=true
GTP_CDR_PATH="cdr/GTP/"
```

#### AI Features
```bash
# Flow Reconstruction
FLOW_RECONSTRUCTION_ENABLED=true
FLOW_DETECT_DEVIATIONS=true
FLOW_STANDARD_PROCEDURES="4G_Attach,5G_Registration,PDU_Session,GTP_Create,MAP_Update_Location"

# Subscriber Correlation
SUBSCRIBER_CORRELATION_ENABLED=true
SUBSCRIBER_TRACK_LOCATION=true
SUBSCRIBER_TRACK_SESSIONS=true
SUBSCRIBER_TIMELINE_ENABLED=true

# AI Analysis
AI_ANALYSIS_ENABLED=true
AI_ANOMALY_DETECTION=true
AI_ROOT_CAUSE_ANALYSIS=true
AI_RECOMMENDATIONS=true
```

---

### 4. system.cfg - System Configuration

**Purpose**: System-wide parameters and performance settings

**Key Parameters:**

#### Network Capture
```bash
CAPTURE_INTERFACE="eth1"               # SPAN/mirror port interface
CAPTURE_FILTER="tcp or udp or sctp"    # BPF filter
CAPTURE_SNAPLEN=65535                  # Snapshot length
CAPTURE_PROMISC=true                   # Promiscuous mode
CAPTURE_TIMEOUT=1000                   # Packet timeout (ms)
BUFFER_SIZE_MB=1024                    # Capture buffer size
```

#### Performance
```bash
WORKER_THREADS=8                       # Number of worker threads
QUEUE_SIZE=10000                       # Message queue size
BATCH_SIZE=100                         # Batch processing size
CACHE_SIZE_MB=512                      # Redis cache size
```

#### Redis Configuration
```bash
REDIS_ENABLED=true
REDIS_HOST="localhost"
REDIS_PORT=6379
REDIS_PASSWORD=""                      # Redis password (if any)
REDIS_DB=0                             # Redis database number
REDIS_POOL_SIZE=20                     # Connection pool size
```

---

### 5. trace.cfg - Logging and Tracing Configuration

**Purpose**: Configure logging levels, rotation, and retention

**Key Parameters:**

#### Log Levels
```bash
LOG_LEVEL="info"                       # panic|fatal|error|warn|info|debug|trace
DEBUG_LOG_ENABLED=false                # Enable debug logging
ERROR_LOG_ENABLED=true                 # Enable error logging
ACCESS_LOG_ENABLED=true                # Enable web access logging
AUDIT_LOG_ENABLED=true                 # Enable audit logging
```

#### Log Files
```bash
APP_LOG_FILE="logs/application/protei-monitoring.log"
ERROR_LOG_FILE="logs/error/error.log"
ACCESS_LOG_FILE="logs/access/access.log"
DEBUG_LOG_FILE="logs/debug/debug.log"
AUDIT_LOG_FILE="logs/system/audit.log"
```

#### Log Rotation
```bash
LOG_ROTATION_SIZE_MB=100               # Rotate when file reaches size
LOG_ROTATION_MAX_FILES=10              # Keep max number of old files
LOG_ROTATION_MAX_AGE_DAYS=30           # Delete files older than
LOG_COMPRESS_OLD_LOGS=true             # Compress rotated logs
```

#### Log Retention
```bash
LOG_RETENTION_DAYS=30                  # General logs
CDR_RETENTION_DAYS=90                  # CDR files
DEBUG_RETENTION_DAYS=7                 # Debug logs
AUDIT_RETENTION_DAYS=365               # Audit logs
```

---

### 6. paths.cfg - File Paths Configuration

**Purpose**: Define all file and directory paths

**Key Parameters:**
```bash
# Installation paths
INSTALL_DIR="/usr/protei/Protei_Monitoring"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/config"
LOG_DIR="$INSTALL_DIR/logs"
CDR_DIR="$INSTALL_DIR/cdr"
TMP_DIR="$INSTALL_DIR/tmp"
LIB_DIR="$INSTALL_DIR/lib"

# Binary paths
MAIN_BINARY="$BIN_DIR/protei-monitoring"
PID_FILE="$TMP_DIR/protei-monitoring.pid"
LOCK_FILE="$TMP_DIR/protei-monitoring.lock"

# CDR paths (per protocol)
CDR_MAP_DIR="$CDR_DIR/MAP"
CDR_CAP_DIR="$CDR_DIR/CAP"
CDR_DIAMETER_DIR="$CDR_DIR/Diameter"
CDR_GTP_DIR="$CDR_DIR/GTP"
CDR_COMBINED_DIR="$CDR_DIR/combined"
```

---

### 7. security.cfg - Security Configuration

**Purpose**: Security settings, authentication, and access control

**Key Parameters:**

#### Authentication
```bash
AUTH_METHOD="local"                    # local|ldap|ad|both
SESSION_TIMEOUT=3600                   # Session timeout (seconds)
TOKEN_EXPIRY=86400                     # JWT token expiry (seconds)
REFRESH_TOKEN_ENABLED=true             # Enable refresh tokens
REFRESH_TOKEN_EXPIRY=604800            # Refresh token expiry (7 days)
```

#### Password Policy
```bash
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBERS=true
PASSWORD_REQUIRE_SPECIAL=true
PASSWORD_HISTORY_COUNT=5               # Prevent password reuse
PASSWORD_EXPIRY_DAYS=90                # Force password change
```

#### Account Lockout
```bash
MAX_LOGIN_ATTEMPTS=5                   # Lock after failed attempts
LOCKOUT_DURATION=1800                  # Lockout duration (seconds)
LOCKOUT_NOTIFY_ADMIN=true              # Email admin on lockout
```

#### LDAP/AD Integration
```bash
LDAP_ENABLED=false
LDAP_HOST="ldap.company.com"
LDAP_PORT=389
LDAP_BASE_DN="dc=company,dc=com"
LDAP_BIND_DN="cn=admin,dc=company,dc=com"
LDAP_BIND_PASSWORD="<ldap_password>"
LDAP_USER_FILTER="(uid=%s)"
LDAP_GROUP_FILTER="(member=%s)"
LDAP_SSL=false                         # Use LDAPS
LDAP_TLS_VERIFY=true                   # Verify certificates

# AD-specific
AD_DOMAIN="company.com"
AD_NETBIOS_NAME="COMPANY"
```

#### RBAC (Role-Based Access Control)
```bash
RBAC_ENABLED=true
DEFAULT_ROLE="viewer"                  # Default role for new users

# Role definitions
ROLE_ADMIN="admin"                     # Full access
ROLE_OPERATOR="operator"               # Read/Write, no system config
ROLE_VIEWER="viewer"                   # Read-only
```

#### TLS/SSL
```bash
HTTPS_ENABLED=false                    # Enable HTTPS
HTTPS_PORT=8443
TLS_CERT_FILE="/path/to/cert.pem"
TLS_KEY_FILE="/path/to/key.pem"
TLS_MIN_VERSION="1.2"                  # Minimum TLS version
TLS_CIPHER_SUITES="TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
```

#### IP Whitelisting
```bash
IP_WHITELIST_ENABLED=false
IP_WHITELIST="192.168.1.0/24,10.0.0.0/8"  # Allowed IP ranges
IP_BLACKLIST_ENABLED=false
IP_BLACKLIST="1.2.3.4,5.6.7.8"        # Blocked IPs
```

---

## Configuration Best Practices

### 1. Backup Configuration Files

```bash
# Backup all configs
sudo tar -czf /backup/protei-config-$(date +%Y%m%d).tar.gz config/

# Restore from backup
sudo tar -xzf /backup/protei-config-20251114.tar.gz -C /usr/protei/Protei_Monitoring/
```

### 2. Validate Configuration

```bash
# Check syntax (bash-compatible files)
bash -n config/license.cfg
bash -n config/db.cfg
bash -n config/protocols.cfg
bash -n config/system.cfg
bash -n config/trace.cfg
bash -n config/paths.cfg
bash -n config/security.cfg

# All should return no errors
```

### 3. Reload Without Downtime

```bash
# Edit configuration
sudo nano config/protocols.cfg

# Reload (no restart required)
sudo ../scripts/reload

# Check logs
tail -f ../logs/application/protei-monitoring.log
```

### 4. Security Checklist

```bash
# Set proper file permissions
sudo chmod 640 config/*.cfg
sudo chown root:protei config/*.cfg

# Protect sensitive files
sudo chmod 600 config/db.cfg          # Contains DB password
sudo chmod 600 config/security.cfg    # Contains LDAP password
sudo chmod 600 config/license.cfg     # Contains license key
```

---

## Configuration Templates

All configuration files use shell variable syntax:
```bash
KEY=value                              # String
KEY="value with spaces"                # Quoted string
KEY=true                               # Boolean
KEY=123                                # Number
KEY="value1,value2,value3"             # Comma-separated list
```

### Environment Variable Override

You can override configuration via environment variables:
```bash
# Override database host
export DB_HOST="db.company.com"

# Override log level
export LOG_LEVEL="debug"

# Start with overrides
sudo -E ../scripts/start
```

---

## Troubleshooting

### Configuration Not Loading

**Issue**: Changes not taking effect

**Solution:**
```bash
# 1. Validate syntax
bash -n config/system.cfg

# 2. Reload configuration
sudo ../scripts/reload

# 3. Check logs for errors
tail -f ../logs/error/error.log

# 4. Restart if reload doesn't work
sudo ../scripts/restart
```

### Database Connection Failed

**Issue**: Cannot connect to database

**Solution:**
```bash
# 1. Verify PostgreSQL is running
sudo systemctl status postgresql

# 2. Test connection manually
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# 3. Check credentials in db.cfg
cat config/db.cfg | grep DB_

# 4. Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

### License Validation Failed

**Issue**: License errors on startup

**Solution:**
```bash
# 1. Check license file
cat config/license.cfg

# 2. Verify MAC address
ip link show | grep ether

# 3. Update LICENSE_MAC in license.cfg

# 4. Check expiry date
date                           # Current date
grep LICENSE_EXPIRY config/license.cfg
```

---

## See Also

- [Installation Guide](../document/INSTALLATION_GUIDE.md)
- [Configuration Guide](../document/CONFIGURATION_GUIDE.md)
- [Administrator Manual](../document/ADMIN_MANUAL.md)
- [Troubleshooting Guide](../document/TROUBLESHOOTING.md)
