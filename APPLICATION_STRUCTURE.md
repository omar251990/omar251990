# Protei Monitoring - Secure Application Structure

## Overview

This document describes the complete secure application structure for Protei Monitoring deployment. The structure follows enterprise security best practices with source code encryption, MAC address binding, and comprehensive configuration management.

---

## Directory Structure

All application files are organized under `/usr/protei/Protei_Monitoring/`:

```
/usr/protei/Protei_Monitoring/
│
├── bin/                # Encrypted binaries and executables
│   └── protei-monitoring     # Main application binary (encrypted)
│
├── config/             # All configuration files
│   ├── license.cfg           # License with MAC binding
│   ├── db.cfg                # Database connection settings
│   ├── protocols.cfg         # Protocol enable/disable
│   ├── system.cfg            # System parameters
│   ├── trace.cfg             # CDR and trace settings
│   ├── paths.cfg             # Directory paths
│   ├── security.cfg          # LDAP, passwords, security
│   └── certs/                # SSL/TLS certificates
│
├── lib/                # Libraries and dependencies
│   ├── decoders/             # Protocol decoders
│   ├── dictionaries/         # Vendor dictionaries
│   │   ├── ericsson/
│   │   ├── huawei/
│   │   ├── zte/
│   │   └── nokia/
│   ├── ai_models/            # AI/ML models
│   └── migrations/           # Database migration scripts
│
├── cdr/                # CDR output (Call Detail Records)
│   ├── MAP/                  # MAP protocol CDRs
│   ├── CAP/                  # CAP protocol CDRs
│   ├── INAP/                 # INAP protocol CDRs
│   ├── Diameter/             # Diameter CDRs
│   ├── GTP/                  # GTP CDRs
│   ├── PFCP/                 # PFCP CDRs
│   ├── HTTP/                 # HTTP/5G SBI CDRs
│   ├── NGAP/                 # NGAP CDRs
│   ├── S1AP/                 # S1AP CDRs
│   ├── NAS/                  # NAS CDRs
│   ├── 2G/                   # 2G network CDRs
│   ├── 3G/                   # 3G network CDRs
│   ├── 4G/                   # 4G network CDRs
│   ├── 5G/                   # 5G network CDRs
│   ├── Roaming/              # Roaming traffic CDRs
│   ├── International/        # International CDRs
│   └── Local/                # Local traffic CDRs
│
├── logs/               # Application logs
│   ├── system/               # System logs
│   ├── error/                # Error logs
│   ├── warning/              # Warning logs
│   ├── trace/                # Protocol trace logs
│   ├── audit/                # Audit logs (user actions)
│   ├── ai/                   # AI analysis logs
│   ├── security/             # Security logs
│   ├── api/                  # API access logs
│   └── debug/                # Debug logs
│
├── scripts/            # Control and utility scripts
│   ├── start*                # Start application
│   ├── stop*                 # Stop application
│   ├── restart*              # Restart application
│   ├── reload*               # Reload configuration
│   ├── status*               # Show status
│   ├── version*              # Show version
│   └── utils/                # Utility scripts
│
├── tmp/                # Temporary files
│   ├── protei.pid            # Process ID file
│   ├── pcap/                 # Temporary PCAP files
│   ├── sessions/             # Session temp data
│   └── processing/           # Processing buffers
│
├── manuals/            # Documentation
│   ├── installation/         # Installation guides
│   ├── configuration/        # Configuration guides
│   ├── operation/            # Operation manuals
│   ├── troubleshooting/      # Troubleshooting guides
│   └── api/                  # API documentation
│
├── pcap/               # PCAP capture files
│   ├── incoming/             # Watch directory for new PCAPs
│   ├── processed/            # Successfully processed
│   ├── failed/               # Failed to process
│   └── archive/              # Archived PCAPs
│
├── data/               # Application data
│   ├── sessions/             # Session data
│   ├── cache/                # Cache files
│   ├── db_backups/           # Database backups
│   └── grafana/              # Grafana dashboards
│
├── output/             # Generated output
│   ├── diagrams/             # Ladder diagrams
│   ├── reports/              # Generated reports
│   └── exports/              # Data exports
│
├── backups/            # Backups
│   ├── database/             # DB backups
│   ├── config/               # Config backups
│   ├── cdr/                  # CDR backups
│   └── logs/                 # Log backups
│
└── VERSION             # Version information file
```

---

## Control Scripts

### Located in `/usr/protei/Protei_Monitoring/scripts/`

All scripts are executable and perform comprehensive validation:

### 1. `start` - Start Application

**Features**:
- ✅ Validates license file and expiry date
- ✅ Verifies MAC address binding
- ✅ Checks enabled protocols from configuration
- ✅ Tests database connectivity
- ✅ Creates all necessary directories
- ✅ Sets proper permissions
- ✅ Starts application with proper environment
- ✅ Verifies startup success

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/start
```

**Checks Performed**:
1. User validation (protei user or root)
2. Checks if already running
3. License validation (expiry, MAC binding)
4. Protocol configuration loading
5. Database connection test
6. Directory creation
7. Process startup
8. Port listening verification

### 2. `stop` - Stop Application

**Features**:
- ✅ Graceful shutdown with SIGTERM
- ✅ Waits for clean exit (30s timeout)
- ✅ Forces kill if needed (SIGKILL)
- ✅ Cleanup of PID file and temp files
- ✅ Logs shutdown event

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/stop
```

**Process**:
1. Find process ID from PID file
2. Send graceful shutdown signal (SIGTERM)
3. Wait up to 30 seconds for clean exit
4. Force kill if timeout
5. Clean up temporary files
6. Remove PID file
7. Log shutdown

### 3. `restart` - Restart Application

**Features**:
- ✅ Stops application cleanly
- ✅ Waits 2 seconds
- ✅ Starts application fresh

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/restart
```

### 4. `reload` - Reload Configuration

**Features**:
- ✅ Reloads configuration without downtime
- ✅ Validates config files before reload
- ✅ Sends SIGHUP signal to application
- ✅ Verifies application still running after reload
- ✅ Logs reload event

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/reload
```

**What Gets Reloaded**:
- Protocol configuration
- Database settings
- Trace/CDR settings
- Path configurations

**Note**: Some changes require full restart:
- License changes
- Server port changes
- TLS certificate changes

### 5. `status` - Show Status

**Features**:
- ✅ Process status (running/not running)
- ✅ Process ID and uptime
- ✅ CPU and memory usage
- ✅ Thread count and file descriptors
- ✅ Network ports listening
- ✅ Active connections count
- ✅ Enabled protocols list
- ✅ License status and expiry
- ✅ Version information
- ✅ Recent log activity

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/status
```

**Output Sections**:
1. Process Status (PID, uptime, start time)
2. Resource Usage (CPU%, Memory%, threads, FDs)
3. Network (listening ports, connections)
4. Enabled Protocols (✓ or ✗ for each)
5. License Status (expiry, days left, customer)
6. Version Information
7. Recent Log Activity

### 6. `version` - Show Version

**Features**:
- ✅ Version and build information
- ✅ Protocol support list
- ✅ Features list
- ✅ License information
- ✅ System information
- ✅ Support contact information

**Usage**:
```bash
/usr/protei/Protei_Monitoring/scripts/version
```

---

## Configuration Files

### Located in `/usr/protei/Protei_Monitoring/config/`

All configuration files use `.cfg` format with bash-compatible syntax.

### 1. `license.cfg` - License Configuration

**Purpose**: License validation, feature enablement, MAC binding

**Key Settings**:
```bash
LICENSE_CUSTOMER="Your Company Name"
LICENSE_EXPIRY="2030-12-31"
LICENSE_SIGNATURE="..."
LICENSE_MAC="00:00:00:00:00:00"  # Hardware lock
LICENSE_MAX_SUBSCRIBERS=10000000
LICENSE_MAX_TPS=10000
```

**Features Controlled**:
- Generation support (2G, 3G, 4G, 5G)
- Protocol enablement (MAP, CAP, Diameter, etc.)
- Advanced features (AI analysis, ML, distributed)

**Security**:
- MAC address binding prevents unauthorized server deployment
- Cryptographic signature validation
- Expiry date enforcement
- Checked on every application start

### 2. `db.cfg` - Database Configuration

**Purpose**: PostgreSQL connection and settings

**Key Settings**:
```bash
DB_ENABLED=true
DB_HOST="localhost"
DB_PORT=5432
DB_NAME="protei_monitoring"
DB_USER="protei_user"
DB_PASSWORD="CHANGE_ME"
DB_MAX_CONNECTIONS=100
DB_AUTO_MIGRATE=true
```

**Features**:
- Connection pooling
- SSL/TLS support
- Auto-migration with Liquibase
- Query timeout settings
- Slow query logging
- Automatic backups

### 3. `protocols.cfg` - Protocol Configuration

**Purpose**: Enable/disable protocols and configure protocol-specific settings

**Protocols Supported**:
- MAP (Mobile Application Part)
- CAP (CAMEL Application Part)
- INAP (Intelligent Network Application Part)
- Diameter (all applications)
- GTP-C v1/v2
- PFCP
- HTTP/2 (5G SBI)
- NGAP (5G)
- S1AP (4G)
- NAS (4G/5G)

**Per-Protocol Settings**:
```bash
MAP_ENABLED=true
MAP_VERSION=3
MAP_DECODE_FULL=true
MAP_SAVE_CDR=true
```

**Analysis Features**:
- Flow reconstruction
- Subscriber correlation
- AI analysis
- Anomaly detection

### 4. `system.cfg` - System Configuration

**Purpose**: General system parameters, server settings

**Major Sections**:
- Server settings (host, port, TLS)
- Metrics server configuration
- WebSocket settings
- JWT authentication
- Session management
- Password policy
- RBAC roles
- Performance settings (threads, buffers, memory)
- Feature flags
- Health checks
- Monitoring intervals

**Security Settings**:
- JWT secret and expiry
- Session timeout
- Password requirements
- Login attempt limits
- Lockout duration

### 5. `trace.cfg` - Trace and CDR Configuration

**Purpose**: CDR generation and trace logging settings

**CDR Settings**:
- CDR format (CSV, JSON, or both)
- CDR fields to include
- CDR organization (by protocol, network, generation)
- File rotation (hourly, daily, weekly)
- Compression settings
- Retention days
- Per-protocol CDR paths

**Trace Settings**:
- Protocol-level tracing
- Trace detail level
- IMSI/procedure filtering
- PCAP capture settings

**Streaming**:
- Real-time CDR streaming to Kafka/RabbitMQ
- Topic configuration

### 6. `paths.cfg` - Paths Configuration

**Purpose**: Custom paths for all directories

**Path Categories**:
- Base directories (bin, config, lib, scripts)
- Log directories (by type)
- CDR directories (by protocol, network, traffic type)
- PCAP directories
- Data and database directories
- Vendor dictionary paths
- Output directories (diagrams, reports)
- Temporary directories
- Backup directories
- Documentation directories

**File Patterns**:
- Log file naming patterns
- CDR file naming patterns
- PCAP file naming patterns

### 7. `security.cfg` - Security Configuration

**Purpose**: LDAP/AD integration, password policies, security features

**LDAP/AD Settings**:
- LDAP server configuration
- Bind credentials
- Search settings
- Attribute mapping
- Group-to-role mapping
- Nested groups support
- Auto-sync settings
- Kerberos/SSO

**Password Policy**:
- Complexity requirements
- Password history
- Age limits
- Common password prevention

**Account Security**:
- Lockout policy
- Failed login protection
- Account deactivation rules

**API Security**:
- API key authentication
- CORS settings
- Rate limiting
- IP whitelist/blacklist

**TLS/SSL**:
- TLS version requirements
- Cipher suites
- Client certificate auth

**Audit & Compliance**:
- Audit logging configuration
- Compliance mode
- Event tracking

**Encryption**:
- Data encryption at rest
- Source code protection
- Binary obfuscation

**Intrusion Detection**:
- Anomaly detection
- Auto-blocking
- Alert configuration

---

## Security Features

### 1. Source Code Encryption

**Protection Method**:
- All source code is encrypted with AES-256-CBC
- Only encrypted binaries are deployed to production
- Application decrypts code in memory only
- No plain-text source code exists on server

**Implementation**:
```bash
SOURCE_CODE_ENCRYPTED=true
SOURCE_CODE_DECRYPT_ON_MEMORY=true
SOURCE_CODE_ALLOW_DUMP=false
```

### 2. MAC Address Binding

**How It Works**:
1. License file contains server MAC address
2. Application reads server's actual MAC on startup
3. Compares against license MAC
4. Refuses to start if mismatch

**Purpose**:
- Prevents copying application to unauthorized servers
- Hardware-locked license
- Protects intellectual property

**Check in start script**:
```bash
validate_mac() {
    LICENSE_MAC="00:11:22:33:44:55"
    CURRENT_MACS=$(ip link show | grep 'link/ether' | awk '{print $2}')
    if ! echo "$CURRENT_MACS" | grep -q "$LICENSE_MAC"; then
        echo "MAC address mismatch - license invalid for this server"
        exit 1
    fi
}
```

### 3. Protocol-Level Security

**Controlled by License**:
- Each protocol can be enabled/disabled via license
- Application validates license before loading decoders
- Disabled protocols cannot be activated

**Runtime Control**:
```bash
# In protocols.cfg
MAP_ENABLED=true
CAP_ENABLED=false  # Cannot enable if license doesn't allow
```

### 4. No External Dependencies

**Self-Contained Design**:
- All libraries in `lib/` directory
- No OS package installation required
- Portable between servers
- No internet connectivity needed for operation

**Benefits**:
- Works on air-gapped systems
- No dependency conflicts
- Consistent behavior across deployments
- Enhanced security (no external dependencies)

### 5. Comprehensive Logging

**Log Categories**:
- System logs (startup, shutdown, errors)
- Security logs (license checks, login attempts)
- Audit logs (configuration changes, user actions)
- Protocol trace logs (message decoding)
- AI analysis logs (decisions, recommendations)

**Retention & Rotation**:
- Automatic log rotation (daily/weekly)
- Configurable retention periods
- Compression for archived logs
- Secure deletion of old logs

---

## Installation Process

### Prerequisites

- Linux OS (Ubuntu 20.04+, CentOS 8+, Debian 11+, RHEL 8+, Rocky Linux 8+)
- Root or sudo access
- Minimum 4 GB RAM, 2 CPU cores, 50 GB disk

### Automated Installation

1. **Build Encrypted Package** (on development machine):
   ```bash
   cd /home/user/omar251990/scripts
   ./build_deployment_package.sh
   ```

2. **Transfer to Server**:
   ```bash
   scp dist/protei-monitoring-*.enc root@SERVER:/tmp/
   ```

3. **Decrypt on Server**:
   ```bash
   cd /tmp
   openssl enc -aes-256-cbc -d -pbkdf2 -iter 100000 \
       -in protei-monitoring-*.enc \
       -out protei-monitoring.tar.gz \
       -k YOUR_ENCRYPTION_KEY
   tar -xzf protei-monitoring.tar.gz
   ```

4. **Run Installer**:
   ```bash
   cd scripts
   chmod +x install.sh
   ./install.sh
   ```

### What the Installer Does

1. ✅ Detects OS and package manager
2. ✅ Updates system packages
3. ✅ Installs dependencies (Go, PostgreSQL, Redis)
4. ✅ Creates `protei` user
5. ✅ Creates all directories under `/usr/protei/Protei_Monitoring/`
6. ✅ Sets proper permissions (700 for bin, 750 for config, etc.)
7. ✅ Deploys binary to `bin/`
8. ✅ Copies configuration templates to `config/`
9. ✅ Generates random passwords for database and admin
10. ✅ Creates database and schema
11. ✅ Generates SSL certificates
12. ✅ Creates systemd service
13. ✅ Configures firewall
14. ✅ Sets packet capture capabilities
15. ✅ Starts application

### Post-Installation

**Access Application**:
```
Web Interface: https://SERVER_IP:8443
Admin Credentials: /etc/protei-monitoring/admin_credentials.txt
```

**Verify Installation**:
```bash
/usr/protei/Protei_Monitoring/scripts/status
```

---

## Operation

### Starting the Application

```bash
/usr/protei/Protei_Monitoring/scripts/start
```

### Stopping the Application

```bash
/usr/protei/Protei_Monitoring/scripts/stop
```

### Restarting the Application

```bash
/usr/protei/Protei_Monitoring/scripts/restart
```

### Reloading Configuration

```bash
/usr/protei/Protei_Monitoring/scripts/reload
```

### Checking Status

```bash
/usr/protei/Protei_Monitoring/scripts/status
```

### Viewing Logs

```bash
# System logs
tail -f /usr/protei/Protei_Monitoring/logs/system/startup.log

# Error logs
tail -f /usr/protei/Protei_Monitoring/logs/error/error.log

# Audit logs
tail -f /usr/protei/Protei_Monitoring/logs/audit/audit.log
```

### Viewing CDRs

```bash
# MAP CDRs
ls -lah /usr/protei/Protei_Monitoring/cdr/MAP/

# 4G CDRs
ls -lah /usr/protei/Protei_Monitoring/cdr/4G/

# Latest CDR
tail /usr/protei/Protei_Monitoring/cdr/MAP/*.csv
```

---

## Troubleshooting

### Application Won't Start

**Check**:
```bash
# License status
cat /usr/protei/Protei_Monitoring/config/license.cfg

# MAC address
ip link show | grep 'link/ether'

# Logs
tail -50 /usr/protei/Protei_Monitoring/logs/system/startup.log
tail -50 /usr/protei/Protei_Monitoring/logs/error/error.log
```

**Common Issues**:
1. License expired
2. MAC address mismatch
3. Database not accessible
4. Port already in use
5. Permissions incorrect

### Database Connection Failed

```bash
# Test database
psql -h localhost -U protei_user -d protei_monitoring

# Check PostgreSQL running
systemctl status postgresql

# Check credentials
cat /usr/protei/Protei_Monitoring/config/db.cfg
```

### Cannot Capture Packets

```bash
# Check capabilities
getcap /usr/protei/Protei_Monitoring/bin/protei-monitoring

# Should show: cap_net_raw,cap_net_admin+eip

# Re-apply if missing
setcap 'cap_net_raw,cap_net_admin+eip' /usr/protei/Protei_Monitoring/bin/protei-monitoring
```

### Web Interface Not Accessible

```bash
# Check if listening
netstat -tlnp | grep 8443

# Check firewall
firewall-cmd --list-ports  # CentOS/RHEL
ufw status                 # Ubuntu

# Open port
firewall-cmd --permanent --add-port=8443/tcp && firewall-cmd --reload
ufw allow 8443/tcp
```

---

## Backup & Restore

### Backup

**Database**:
```bash
sudo -u postgres pg_dump protei_monitoring > /backup/db_$(date +%Y%m%d).sql
```

**Configuration**:
```bash
tar -czf /backup/config_$(date +%Y%m%d).tar.gz /usr/protei/Protei_Monitoring/config/
```

**CDR Files**:
```bash
tar -czf /backup/cdr_$(date +%Y%m%d).tar.gz /usr/protei/Protei_Monitoring/cdr/
```

### Restore

**Database**:
```bash
sudo -u postgres psql protei_monitoring < /backup/db_20250114.sql
```

**Configuration**:
```bash
tar -xzf /backup/config_20250114.tar.gz -C /
```

---

## Maintenance

### Log Rotation

Logs are automatically rotated based on `trace.cfg` settings:
- System logs: Daily rotation, 30 days retention
- CDR files: Daily rotation, 90 days retention, compressed
- PCAP files: Hourly rotation, 7 days retention

### Database Maintenance

**Vacuum**:
```bash
vacuumdb -U protei_user -d protei_monitoring
```

**Analyze**:
```bash
vacuumdb -U protei_user -d protei_monitoring --analyze
```

### CDR Cleanup

Old CDRs are automatically cleaned based on retention settings in `trace.cfg`:
```bash
CDR_RETENTION_DAYS=90
CDR_AUTO_DELETE_OLD=true
```

---

## Upgrades

### Upgrade Process

1. **Stop application**:
   ```bash
   /usr/protei/Protei_Monitoring/scripts/stop
   ```

2. **Backup current installation**:
   ```bash
   cp -a /usr/protei/Protei_Monitoring /usr/protei/Protei_Monitoring.backup
   ```

3. **Decrypt new package**

4. **Copy new binary**:
   ```bash
   cp bin/protei-monitoring /usr/protei/Protei_Monitoring/bin/
   ```

5. **Update configuration if needed**

6. **Run database migrations**

7. **Start application**:
   ```bash
   /usr/protei/Protei_Monitoring/scripts/start
   ```

8. **Verify**:
   ```bash
   /usr/protei/Protei_Monitoring/scripts/status
   ```

---

## Security Best Practices

1. **Change Default Passwords**:
   - Admin password
   - Database password
   - JWT secret

2. **Use Valid SSL Certificates**:
   - Replace self-signed certificates
   - Obtain from trusted CA

3. **Restrict Access**:
   - Configure firewall rules
   - Use IP whitelist
   - Enable RBAC

4. **Regular Backups**:
   - Daily database backups
   - Configuration backups
   - CDR backups

5. **Monitor Logs**:
   - Check audit logs regularly
   - Monitor security logs
   - Set up alerts for errors

6. **Keep Encryption Key Safe**:
   - Store in password manager
   - Create multiple backups
   - Never commit to version control

7. **Update Regularly**:
   - Apply security patches
   - Update dependencies
   - Test upgrades in staging first

---

## Support

For technical support:
- Email: support@protei-monitoring.com
- Documentation: https://docs.protei-monitoring.com

For licensing:
- Email: sales@protei-monitoring.com

---

## License

Copyright © 2025 Protei Monitoring. All rights reserved.

This software is proprietary and confidential. Unauthorized copying, distribution, or use is strictly prohibited.

All intellectual property rights are owned by Protei Monitoring.
