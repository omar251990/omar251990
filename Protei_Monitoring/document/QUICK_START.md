# Protei Monitoring v2.0 - Quick Start Guide

Get started with Protei Monitoring in **5 minutes**!

## Prerequisites

- Linux server (RHEL/CentOS/Ubuntu)
- PostgreSQL 14+ installed and running
- Redis 6+ installed and running
- Root or sudo access
- Valid license file

---

## Step 1: Extract and Install (2 minutes)

```bash
# Extract package
cd /home/user/omar251990
sudo tar -xzf Protei_Monitoring-2.0.0.tar.gz -C /usr/protei/

# Navigate to installation directory
cd /usr/protei/Protei_Monitoring
```

---

## Step 2: Basic Configuration (2 minutes)

### Configure License

```bash
# Edit license configuration
sudo nano config/license.cfg
```

Update these two lines:
```bash
LICENSE_MAC="<your_server_mac_address>"  # Get from: ip link show
LICENSE_EXPIRY="2026-12-31"              # Your license expiry date
```

### Configure Database

```bash
# Edit database configuration
sudo nano config/db.cfg
```

Update connection details:
```bash
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="protei_monitoring"
DB_USER="protei"
DB_PASSWORD="<your_db_password>"
```

### Create Database

```bash
# Create database and user
sudo -u postgres psql <<EOF
CREATE DATABASE protei_monitoring;
CREATE USER protei WITH PASSWORD '<your_db_password>';
GRANT ALL PRIVILEGES ON DATABASE protei_monitoring TO protei;
\c protei_monitoring
GRANT ALL ON SCHEMA public TO protei;
EOF
```

---

## Step 3: Start Application (1 minute)

```bash
# Make scripts executable
sudo chmod +x scripts/*

# Start the application
sudo scripts/start
```

**Expected Output:**
```
📚 Initializing knowledge base...
  ✅ Loaded 18 standards and 10 protocols
🤖 Initializing AI analysis engine...
  ✅ AI analysis engine ready (7 detection rules)
🔄 Initializing flow reconstructor...
  ✅ Flow reconstructor ready (5 procedure templates)
👤 Initializing subscriber correlator...
  ✅ Subscriber correlator ready (multi-identifier tracking)
🌐 Initializing web server...
  ✅ Web server ready on port 8080
✅ Protei Monitoring started successfully (PID: 12345)
```

---

## Step 4: Access Web Interface (30 seconds)

1. **Open your browser:**
   ```
   http://<server_ip>:8080
   ```

2. **Login with default credentials:**
   - Username: `admin`
   - Password: `admin` (change immediately!)

3. **You should see the dashboard** with:
   - System status
   - Active protocols (10 total)
   - Real-time statistics
   - Recent sessions

---

## Quick Verification

### Check Application Status

```bash
sudo scripts/status
```

### Test API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# List protocols
curl http://localhost:8080/api/protocols

# Knowledge base standards
curl http://localhost:8080/api/knowledge/standards

# AI analysis issues
curl http://localhost:8080/api/analysis/issues
```

### View Logs

```bash
# Application log
tail -f logs/application/protei-monitoring.log

# Error log
tail -f logs/error/error.log
```

---

## What's Next?

### For End Users

1. **Explore the Dashboard**
   - View real-time traffic statistics
   - Monitor active sessions
   - Check protocol health

2. **Search and Filter**
   - Search by IMSI, MSISDN, or IMEI
   - Filter by protocol, time range
   - Export CDRs

3. **Use AI Features**
   - View detected issues
   - Check subscriber timelines
   - Analyze message flows

### For Administrators

1. **Configure Network Capture**
   ```bash
   sudo nano config/system.cfg
   # Set CAPTURE_INTERFACE to your SPAN/mirror port
   ```

2. **Set Up Users and Roles**
   - Go to: Settings → User Management
   - Create users with appropriate roles
   - Configure LDAP (optional)

3. **Enable Monitoring**
   - Set up log rotation
   - Configure alerting thresholds
   - Enable health checks

4. **Security Hardening**
   ```bash
   # Change default admin password
   # Enable HTTPS
   # Configure firewall rules
   # Review security.cfg
   ```

---

## Common Commands

```bash
# Start application
sudo scripts/start

# Stop application
sudo scripts/stop

# Restart application
sudo scripts/restart

# Reload configuration (no downtime)
sudo scripts/reload

# Check status
sudo scripts/status

# View version
sudo scripts/version
```

---

## Quick Troubleshooting

### Application Won't Start

```bash
# Check license
cat config/license.cfg

# Verify MAC address matches
ip link show | grep ether

# Check database connection
psql -h localhost -U protei -d protei_monitoring -c "SELECT 1;"
```

### Cannot Access Web UI

```bash
# Check application is running
sudo scripts/status

# Verify port is listening
sudo netstat -tulpn | grep 8080

# Check firewall
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

### No Traffic Visible

```bash
# Verify capture interface
sudo tcpdump -i eth1 -c 10

# Check protocol configuration
cat config/protocols.cfg | grep ENABLED

# Review logs
tail -f logs/debug/debug.log
```

---

## Default Ports

- **8080**: Web UI (HTTP)
- **8443**: Web UI (HTTPS, if configured)
- **5432**: PostgreSQL database
- **6379**: Redis cache

---

## Important Files

```
/usr/protei/Protei_Monitoring/
├── config/
│   ├── license.cfg      # License configuration
│   ├── db.cfg          # Database settings
│   ├── protocols.cfg   # Protocol enable/disable
│   └── system.cfg      # System parameters
├── logs/
│   ├── application/    # Application logs
│   └── error/          # Error logs
├── cdr/                # CDR output files
└── scripts/
    ├── start           # Start script
    ├── stop            # Stop script
    └── status          # Status script
```

---

## Getting Help

- **Full Documentation**: See `document/README.md`
- **Installation Issues**: See `INSTALLATION_GUIDE.md`
- **User Manual**: See `USER_MANUAL.md`
- **API Reference**: See `API_REFERENCE.md`
- **Support Email**: support@protei.com

---

## Feature Highlights

### ✅ Protocol Support (10 Protocols)
- MAP, CAP, INAP (2G/3G signaling)
- Diameter (4G/5G signaling)
- GTP, PFCP (4G/5G data plane)
- HTTP/2 (5G SBI)
- NGAP, S1AP (4G/5G RAN)
- NAS (4G/5G mobile)

### ✅ AI & Intelligence
- **Knowledge Base**: 18 3GPP standards built-in
- **AI Analysis**: 7 intelligent detection rules
- **Flow Reconstruction**: 5 standard procedures
- **Subscriber Correlation**: Multi-identifier tracking

### ✅ Web Interface
- Real-time dashboard
- Advanced search and filtering
- Ladder diagram visualization
- CDR export (CSV, JSON, XML)
- User management with RBAC

### ✅ Enterprise Features
- LDAP/AD integration
- MAC address binding
- Secure configuration
- Complete audit logging
- High availability support

---

**You're now ready to use Protei Monitoring!**

For detailed information, explore the full documentation in the `document/` directory.
