# scripts/ - Control Scripts

Application control and utility scripts.

## Main Control Scripts

All scripts should be run with sudo from the Protei_Monitoring directory:

```bash
cd /usr/protei/Protei_Monitoring
sudo scripts/<script_name>
```

### start
**Purpose**: Start the Protei Monitoring application

**Usage:**
```bash
sudo scripts/start
```

**Actions:**
- Validates license file
- Checks MAC address binding
- Verifies database connectivity
- Checks required protocols
- Starts application in background
- Creates PID file
- Returns exit code (0=success, non-zero=failure)

**Output:**
```
📚 Initializing knowledge base...
  ✅ Loaded 18 standards and 10 protocols
🤖 Initializing AI analysis engine...
  ✅ AI analysis engine ready (7 detection rules)
🔄 Initializing flow reconstructor...
  ✅ Flow reconstructor ready (5 procedure templates)
👤 Initializing subscriber correlator...
  ✅ Subscriber correlator ready
🌐 Initializing web server...
  ✅ Web server ready on port 8080
✅ Protei Monitoring started successfully (PID: 12345)
```

---

### stop
**Purpose**: Stop the Protei Monitoring application

**Usage:**
```bash
sudo scripts/stop
```

**Actions:**
- Sends SIGTERM to application process
- Waits for graceful shutdown (up to 30 seconds)
- Sends SIGKILL if not stopped gracefully
- Removes PID file
- Returns exit code

**Output:**
```
🛑 Stopping Protei Monitoring...
✅ Protei Monitoring stopped successfully
```

---

### restart
**Purpose**: Restart the application (stop then start)

**Usage:**
```bash
sudo scripts/restart
```

**Actions:**
- Calls stop script
- Waits 2 seconds
- Calls start script

---

### reload
**Purpose**: Reload configuration without restarting

**Usage:**
```bash
sudo scripts/reload
```

**Actions:**
- Sends SIGHUP to application process
- Application reloads all .cfg files
- No downtime or connection interruption

**What gets reloaded:**
- ✅ Protocol enable/disable
- ✅ Log levels
- ✅ Performance tuning (workers, buffers)
- ✅ Security settings
- ❌ Database connection (requires restart)
- ❌ License file (requires restart)

**Output:**
```
🔄 Reloading configuration...
✅ Configuration reloaded successfully
```

---

### status
**Purpose**: Check application status and health

**Usage:**
```bash
sudo scripts/status
```

**Output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Protei Monitoring - System Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔹 Process Information
Status: RUNNING
PID: 12345
Uptime: 2 days, 5 hours, 32 minutes
User: protei
Started: 2025-11-12 08:15:30

🔹 Resource Usage
CPU: 15.2%
Memory: 2.1 GB / 32.0 GB (6.5%)
Disk: 45.3 GB / 500 GB (9%)
Network: eth1 (SPAN port)

🔹 Protocol Status
✅ MAP      - Active (12,345 sessions)
✅ CAP      - Active (3,456 sessions)
✅ INAP     - Active (1,234 sessions)
✅ Diameter - Active (45,678 sessions)
✅ GTP      - Active (123,456 sessions)
✅ PFCP     - Active (34,567 sessions)
✅ HTTP/2   - Active (8,901 sessions)
✅ NGAP     - Active (12,345 sessions)
✅ S1AP     - Active (9,876 sessions)
✅ NAS      - Active (11,234 sessions)

🔹 License Information
Type: Enterprise
Expiry: 2026-12-31 (395 days remaining)
Protocols: 10/10 enabled
Max Sessions: 10000 (current: 262,772)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### version
**Purpose**: Display version and build information

**Usage:**
```bash
sudo scripts/version
```

**Output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Protei Monitoring - Version Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Application: Protei_Monitoring
Version: 2.0.0
Build Date: 2025-11-14
Build Type: Production
Go Version: go1.21.5
OS/Arch: linux/amd64

🔹 Protocol Support (10 Protocols)
  MAP      - v3  (3GPP TS 29.002)
  CAP      - v4  (3GPP TS 29.078)
  INAP     - CS2 (ITU-T Q.1218)
  Diameter - RFC 6733 (3GPP TS 29.272, 29.273)
  GTP      - v2  (3GPP TS 29.274, 29.281)
  PFCP     - v1  (3GPP TS 29.244)
  HTTP/2   - RFC 7540 (3GPP TS 29.500)
  NGAP     - v1  (3GPP TS 38.413)
  S1AP     - v1  (3GPP TS 36.413)
  NAS      - v1  (3GPP TS 24.301, 24.501)

🔹 AI Features
  Knowledge Base    - 18 standards
  Analysis Engine   - 7 detection rules
  Flow Reconstructor - 5 procedures
  Subscriber Correlation - Multi-identifier

🔹 Installation
Install Path: /usr/protei/Protei_Monitoring
Config Path: /usr/protei/Protei_Monitoring/config
Log Path: /usr/protei/Protei_Monitoring/logs
CDR Path: /usr/protei/Protei_Monitoring/cdr

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Utility Scripts (utils/)

Additional utility scripts for maintenance and troubleshooting.

### utils/backup.sh
**Purpose**: Backup configuration and database

**Usage:**
```bash
sudo scripts/utils/backup.sh [destination]
```

**Creates:**
- Configuration backup (tar.gz)
- Database dump (SQL)
- CDR archive (tar.gz)

---

### utils/restore.sh
**Purpose**: Restore from backup

**Usage:**
```bash
sudo scripts/utils/restore.sh <backup_file>
```

---

### utils/health_check.sh
**Purpose**: Comprehensive health check

**Usage:**
```bash
sudo scripts/utils/health_check.sh
```

**Checks:**
- Application process
- Database connectivity
- Redis connectivity
- Disk space
- Memory usage
- Log errors
- License expiry

---

### utils/analyze_logs.sh
**Purpose**: Analyze logs for issues

**Usage:**
```bash
sudo scripts/utils/analyze_logs.sh [hours]
```

**Reports:**
- Error count by type
- Top error messages
- Slow queries
- Failed sessions
- Security events

---

### utils/export_cdr.sh
**Purpose**: Export CDRs to external system

**Usage:**
```bash
sudo scripts/utils/export_cdr.sh [destination]
```

---

### utils/cleanup.sh
**Purpose**: Clean old logs and temporary files

**Usage:**
```bash
sudo scripts/utils/cleanup.sh [days]
```

**Cleans:**
- Old log files
- Rotated logs
- Temporary files
- Old CDRs (based on retention policy)

---

## Exit Codes

All scripts use standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | License error |
| 4 | Database error |
| 5 | Network error |
| 6 | Permission error |
| 7 | Already running/not running |

**Example:**
```bash
sudo scripts/start
if [ $? -eq 0 ]; then
  echo "Started successfully"
else
  echo "Failed to start"
  exit 1
fi
```

---

## Automation

### Systemd Service

Use systemd for automatic startup:

```bash
sudo systemctl enable protei-monitoring
sudo systemctl start protei-monitoring
sudo systemctl status protei-monitoring
```

### Cron Jobs

Example cron tasks:

```cron
# Daily backup at 2 AM
0 2 * * * /usr/protei/Protei_Monitoring/scripts/utils/backup.sh

# Cleanup old logs weekly
0 3 * * 0 /usr/protei/Protei_Monitoring/scripts/utils/cleanup.sh 30

# Export CDRs daily
0 1 * * * /usr/protei/Protei_Monitoring/scripts/utils/export_cdr.sh

# Health check every hour
0 * * * * /usr/protei/Protei_Monitoring/scripts/utils/health_check.sh
```

---

## Troubleshooting

### Script Fails to Execute

```bash
# Make scripts executable
sudo chmod +x scripts/*
sudo chmod +x scripts/utils/*
```

### Permission Denied

```bash
# Run with sudo
sudo scripts/start

# Or change ownership
sudo chown -R root:protei scripts/
sudo chmod 755 scripts/*.sh
```

### Application Won't Start

```bash
# Check logs
sudo scripts/start
# If fails, check:
tail -f logs/error/error.log

# Verify configuration
bash -n config/*.cfg

# Check license
cat config/license.cfg
```

---

## See Also

- [Administrator Manual](../document/ADMIN_MANUAL.md)
- [Configuration Guide](../document/CONFIGURATION_GUIDE.md)
- [Troubleshooting Guide](../document/TROUBLESHOOTING.md)
- [Deployment Guide](../document/DEPLOYMENT_GUIDE.md)
