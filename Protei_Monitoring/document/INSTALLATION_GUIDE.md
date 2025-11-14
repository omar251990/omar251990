# Protei Monitoring v2.0 - Installation Guide

## Table of Contents
1. [System Requirements](#system-requirements)
2. [Pre-Installation Checklist](#pre-installation-checklist)
3. [Installation Methods](#installation-methods)
4. [Post-Installation Configuration](#post-installation-configuration)
5. [Verification](#verification)
6. [Troubleshooting](#troubleshooting)

---

## System Requirements

### Hardware Requirements

**Minimum Requirements:**
- **CPU**: 4 cores @ 2.4 GHz
- **RAM**: 8 GB
- **Disk**: 100 GB SSD
- **Network**: 1 Gbps NIC

**Recommended for Production:**
- **CPU**: 16 cores @ 3.0 GHz or higher
- **RAM**: 32 GB or more
- **Disk**: 500 GB NVMe SSD (RAID 10 recommended)
- **Network**: 10 Gbps NIC with SPAN/mirror port access

### Software Requirements

**Operating System:**
- RHEL 8.x / 9.x
- CentOS 8.x / 9.x
- Ubuntu 20.04 LTS / 22.04 LTS
- Debian 11 / 12

**Dependencies:**
- **PostgreSQL**: 14.x or 15.x
- **Redis**: 6.x or 7.x
- **Go Runtime**: 1.21+ (for building from source)
- **libpcap**: 1.10+ (for packet capture)

**Network Requirements:**
- Access to SPAN/mirror port or TAP device
- Outbound HTTPS access (for license validation)
- Ports required:
  - **8080**: Web UI (HTTP)
  - **8443**: Web UI (HTTPS)
  - **5432**: PostgreSQL
  - **6379**: Redis

---

## Pre-Installation Checklist

### 1. Verify License File

```bash
# Check license file exists and is valid
cat /path/to/license.json

# Expected fields:
# - customer_name
# - license_key
# - expiry_date
# - max_protocols
# - mac_address (must match server MAC)
```

### 2. Check Network Access

```bash
# Verify SPAN/mirror port access
sudo tcpdump -i eth1 -c 10

# Check database connectivity
pg_isready -h db_server -p 5432

# Check Redis connectivity
redis-cli -h redis_server ping
```

### 3. Verify Permissions

```bash
# Ensure you have root or sudo access
sudo whoami

# Check disk space
df -h /usr/protei

# Verify network capture capability
sudo setcap cap_net_raw,cap_net_admin=eip /path/to/binary
```

---

## Installation Methods

### Method 1: Package Installation (Recommended)

#### Step 1: Extract Package

```bash
# Extract the Protei_Monitoring package
cd /home/user/omar251990
tar -xzf Protei_Monitoring-2.0.0.tar.gz

# Move to installation directory
sudo mv Protei_Monitoring /usr/protei/
cd /usr/protei/Protei_Monitoring
```

#### Step 2: Configure License

```bash
# Copy your license file to config directory
sudo cp /path/to/your/license.json config/license.cfg

# Edit license configuration
sudo nano config/license.cfg

# Update MAC address binding (get from: ip link show)
LICENSE_MAC="xx:xx:xx:xx:xx:xx"
LICENSE_EXPIRY="2026-12-31"
```

#### Step 3: Configure Database

```bash
# Edit database configuration
sudo nano config/db.cfg

# Update connection settings:
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="protei_monitoring"
DB_USER="protei"
DB_PASSWORD="<secure_password>"
```

#### Step 4: Configure Protocols

```bash
# Edit protocol configuration
sudo nano config/protocols.cfg

# Enable desired protocols (default: all enabled)
MAP_ENABLED=true
CAP_ENABLED=true
DIAMETER_ENABLED=true
GTP_ENABLED=true
# ... etc
```

#### Step 5: Run Installation

```bash
# Make scripts executable
sudo chmod +x scripts/*

# Run initial setup (creates database schema, users, etc.)
sudo scripts/install.sh

# Expected output:
# ✅ Database schema created
# ✅ Default admin user created
# ✅ Configuration validated
# ✅ Installation complete
```

#### Step 6: Start Application

```bash
# Start the application
sudo scripts/start

# Expected output:
# 📚 Initializing knowledge base...
#   ✅ Loaded 18 standards and 10 protocols
# 🤖 Initializing AI analysis engine...
#   ✅ AI analysis engine ready (7 detection rules)
# 🔄 Initializing flow reconstructor...
#   ✅ Flow reconstructor ready (5 procedure templates)
# 👤 Initializing subscriber correlator...
#   ✅ Subscriber correlator ready
# 🌐 Initializing web server...
#   ✅ Web server ready on port 8080
# ✅ Protei Monitoring started successfully (PID: xxxxx)
```

### Method 2: Docker Installation

```bash
# Pull Docker image
docker pull protei/monitoring:2.0.0

# Run container
docker run -d \
  --name protei-monitoring \
  --network host \
  --cap-add=NET_RAW \
  --cap-add=NET_ADMIN \
  -v /usr/protei/config:/app/config \
  -v /usr/protei/logs:/app/logs \
  -v /usr/protei/cdr:/app/cdr \
  -e DB_HOST=postgres_host \
  -e REDIS_HOST=redis_host \
  protei/monitoring:2.0.0
```

### Method 3: Build from Source

```bash
# Clone repository
git clone https://github.com/protei/monitoring.git
cd monitoring

# Install Go dependencies
go mod download

# Build application
go build -o protei-monitoring ./cmd/protei-monitoring

# Copy to installation directory
sudo mkdir -p /usr/protei/Protei_Monitoring/bin
sudo cp protei-monitoring /usr/protei/Protei_Monitoring/bin/
```

---

## Post-Installation Configuration

### 1. Create Admin User

```bash
# Access the database
psql -h localhost -U protei -d protei_monitoring

# Create admin user
INSERT INTO users (username, password_hash, role, created_at)
VALUES ('admin', '$2a$10$...', 'admin', NOW());

# Exit
\q
```

### 2. Configure Network Capture

```bash
# Edit system configuration
sudo nano config/system.cfg

# Set capture interface
CAPTURE_INTERFACE="eth1"  # SPAN/mirror port
CAPTURE_FILTER="tcp or udp or sctp"
BUFFER_SIZE_MB=1024
```

### 3. Configure Security

```bash
# Edit security configuration
sudo nano config/security.cfg

# Enable LDAP (optional)
LDAP_ENABLED=true
LDAP_HOST="ldap.company.com"
LDAP_BASE_DN="dc=company,dc=com"

# Configure session timeout
SESSION_TIMEOUT=3600  # 1 hour
MAX_LOGIN_ATTEMPTS=5
```

### 4. Configure Log Retention

```bash
# Edit trace/log configuration
sudo nano config/trace.cfg

# Set retention policies
LOG_RETENTION_DAYS=30
CDR_RETENTION_DAYS=90
DEBUG_LOG_ENABLED=false
```

### 5. Set Up Systemd Service (Optional)

```bash
# Create systemd service file
sudo nano /etc/systemd/system/protei-monitoring.service
```

```ini
[Unit]
Description=Protei Monitoring v2.0
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=protei
Group=protei
WorkingDirectory=/usr/protei/Protei_Monitoring
ExecStart=/usr/protei/Protei_Monitoring/scripts/start
ExecStop=/usr/protei/Protei_Monitoring/scripts/stop
ExecReload=/usr/protei/Protei_Monitoring/scripts/reload
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable protei-monitoring
sudo systemctl start protei-monitoring
```

---

## Verification

### 1. Check Application Status

```bash
# Using control script
sudo scripts/status

# Expected output:
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Protei Monitoring - System Status
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Status: RUNNING
# PID: 12345
# Uptime: 00:05:32
# CPU Usage: 15.2%
# Memory Usage: 2.1 GB / 32.0 GB
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 2. Test Web Interface

```bash
# Test health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","version":"2.0.0","uptime":332}

# Access web UI
firefox http://localhost:8080
# Login: admin / <your_password>
```

### 3. Verify Protocol Decoders

```bash
# Check loaded protocols
curl http://localhost:8080/api/protocols

# Expected: List of 10 protocols with status
```

### 4. Check Logs

```bash
# View application logs
tail -f logs/application/protei-monitoring.log

# Check for errors
grep ERROR logs/error/*.log
```

### 5. Verify Database Connection

```bash
# Check database connection
psql -h localhost -U protei -d protei_monitoring -c "SELECT COUNT(*) FROM sessions;"

# Should return: count (may be 0 initially)
```

---

## Troubleshooting

### Issue: Application Won't Start

**Symptoms:**
- `scripts/start` fails
- Error: "License validation failed"

**Solutions:**

```bash
# 1. Check license file
cat config/license.cfg

# 2. Verify MAC address
ip link show | grep ether
# Update LICENSE_MAC in license.cfg to match

# 3. Check expiry date
date  # Current date
# Ensure LICENSE_EXPIRY is in the future

# 4. Validate license format
bash -n config/license.cfg  # Should return no errors
```

### Issue: Cannot Capture Packets

**Symptoms:**
- No traffic visible in UI
- Error: "Permission denied" on capture

**Solutions:**

```bash
# 1. Grant capture capabilities
sudo setcap cap_net_raw,cap_net_admin=eip bin/protei-monitoring

# 2. Check interface exists
ip link show eth1

# 3. Verify SPAN/mirror configuration on switch

# 4. Test manual capture
sudo tcpdump -i eth1 -c 10
```

### Issue: Database Connection Failed

**Symptoms:**
- Error: "Could not connect to database"

**Solutions:**

```bash
# 1. Check PostgreSQL is running
sudo systemctl status postgresql

# 2. Verify connection settings
psql -h localhost -U protei -d protei_monitoring

# 3. Check firewall
sudo firewall-cmd --list-all | grep 5432

# 4. Review database logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

### Issue: Web UI Not Accessible

**Symptoms:**
- Cannot access http://localhost:8080
- Connection refused

**Solutions:**

```bash
# 1. Check application is running
sudo scripts/status

# 2. Verify port binding
sudo netstat -tulpn | grep 8080

# 3. Check firewall
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload

# 4. Review web server logs
tail -f logs/access/access.log
```

### Issue: High CPU/Memory Usage

**Symptoms:**
- CPU usage > 80%
- Memory usage growing continuously

**Solutions:**

```bash
# 1. Check number of workers
nano config/system.cfg
# Reduce: WORKER_THREADS=4

# 2. Adjust buffer sizes
nano config/protocols.cfg
# Reduce: BUFFER_SIZE_MB=512

# 3. Enable log rotation
nano config/trace.cfg
# Set: LOG_ROTATION_SIZE_MB=100

# 4. Restart application
sudo scripts/restart
```

---

## Next Steps

After successful installation:

1. **Configure Users**: [Administrator Manual](ADMIN_MANUAL.md)
2. **Set Up Monitoring**: [Monitoring & Alerting](MONITORING_ALERTING.md)
3. **Learn the Interface**: [Web Interface Guide](WEB_INTERFACE_GUIDE.md)
4. **Integrate Systems**: [Integration Guide](INTEGRATION_GUIDE.md)

---

## Support

For installation support:
- Email: support@protei.com
- Documentation: https://docs.protei.com/installation
- Emergency: +1-XXX-XXX-XXXX (24/7)
