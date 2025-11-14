# Protei Monitoring - Deployment and Installation Guide

## Overview

This guide explains how to build, encrypt, and deploy the Protei Monitoring application on a production server. The deployment process includes source code encryption to protect your intellectual property.

## Table of Contents

1. [Security Architecture](#security-architecture)
2. [Build Process](#build-process)
3. [Installation on Server](#installation-on-server)
4. [Testing the Deployment](#testing-the-deployment)
5. [Source Code Protection](#source-code-protection)
6. [Troubleshooting](#troubleshooting)

---

## Security Architecture

### Encryption Strategy

The Protei Monitoring deployment uses a **three-layer security approach**:

1. **Binary Compilation**: Source code is compiled into a native binary, making reverse engineering difficult
2. **Symbol Stripping**: Debugging symbols are removed from the binary
3. **Package Encryption**: The entire deployment package is encrypted with AES-256-CBC encryption

### What Gets Encrypted?

- Application binary
- Web interface files (HTML, CSS, JS)
- Configuration templates
- Installation scripts
- Documentation

### Source Code Protection

The source code itself remains on your development machine and is:
- **Never deployed** to production servers (only the compiled binary is deployed)
- **Optionally encrypted** separately if you need to store or transfer it
- **Protected by encryption** when transferred or stored

Only you (with the encryption key) can decrypt and view the source code.

---

## Build Process

### Prerequisites

On your **development machine**, you need:
- Linux OS (Ubuntu, Debian, CentOS, etc.)
- Go 1.21 or higher
- Git (optional, for version info)
- OpenSSL
- Standard build tools (gcc, make)

### Step 1: Prepare Your Environment

```bash
# Ensure you're in the project directory
cd /path/to/protei-monitoring

# Verify Go is installed
go version

# Verify dependencies
go mod download
go mod verify
```

### Step 2: Build Encrypted Deployment Package

Run the build script:

```bash
cd scripts
chmod +x build_deployment_package.sh
./build_deployment_package.sh
```

**The script will:**

1. ✅ Check all dependencies
2. ✅ Clean previous builds
3. ✅ Compile the application binary (optimized)
4. ✅ Strip debugging symbols
5. ✅ Copy necessary files (web assets, configs, scripts)
6. ✅ Create version information
7. ✅ Generate encryption key (or use existing)
8. ✅ Encrypt the package with AES-256-CBC
9. ✅ Generate SHA256 checksum
10. ✅ Create deployment documentation

### Step 3: Output Files

After building, check the `dist/` directory:

```bash
cd ../dist
ls -lh
```

You should see:
- `protei-monitoring-YYYYMMDD_HHMMSS.enc` - Encrypted deployment package
- `protei-monitoring-YYYYMMDD_HHMMSS.enc.sha256` - Checksum file
- `ENCRYPTION_KEY.txt` - **CRITICAL: Your encryption key**
- `DEPLOY.md` - Full deployment instructions
- `QUICKSTART.txt` - Quick reference guide

### Step 4: Secure Your Encryption Key

**CRITICAL STEPS:**

1. **Copy the encryption key** from `ENCRYPTION_KEY.txt`
2. **Store it securely** in multiple locations:
   - Password manager (1Password, LastPass, BitWarden, etc.)
   - Encrypted USB drive (backup)
   - Secure company vault
3. **NEVER** commit the key to version control
4. **NEVER** send the key via unencrypted email or chat
5. **Create backups** - if you lose the key, you lose access to the deployment

---

## Installation on Server

### Server Prerequisites

Your **production server** must have:
- **OS**: Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+, Rocky Linux 8+, Fedora 34+
- **CPU**: 2+ cores (4+ recommended)
- **RAM**: 4 GB minimum (8 GB+ recommended)
- **Disk**: 50 GB minimum (100 GB+ recommended)
- **Network**: Internet connection for dependency installation
- **Access**: Root or sudo privileges

### Automated Installation (Recommended)

#### Step 1: Transfer Package to Server

```bash
# From your development machine
cd dist

# Transfer encrypted package (replace YOUR_SERVER with actual IP/hostname)
scp protei-monitoring-20250114_*.enc root@YOUR_SERVER:/tmp/
```

#### Step 2: Connect to Server

```bash
ssh root@YOUR_SERVER
```

#### Step 3: Decrypt the Package

```bash
cd /tmp

# Decrypt (you'll be prompted for the encryption key)
openssl enc -aes-256-cbc -d -pbkdf2 -iter 100000 \
    -in protei-monitoring-*.enc \
    -out protei-monitoring.tar.gz \
    -k YOUR_ENCRYPTION_KEY
```

**When prompted**, enter your encryption key (from `ENCRYPTION_KEY.txt`).

#### Step 4: Verify Checksum (Optional but Recommended)

```bash
# Calculate checksum
sha256sum protei-monitoring.tar.gz

# Compare with the checksum from your development machine
# They should match!
```

#### Step 5: Extract Package

```bash
tar -xzf protei-monitoring.tar.gz
```

#### Step 6: Run Automated Installer

```bash
cd scripts
chmod +x install.sh
./install.sh
```

**The installer will automatically:**

1. ✅ Detect your operating system
2. ✅ Update system packages
3. ✅ Install Go 1.21.5
4. ✅ Install PostgreSQL 14
5. ✅ Install Redis 7
6. ✅ Create application user (`protei`)
7. ✅ Create directories (`/opt/protei-monitoring`, `/etc/protei-monitoring`, etc.)
8. ✅ Setup PostgreSQL database
9. ✅ Deploy application binary
10. ✅ Create configuration file
11. ✅ Generate SSL certificates (self-signed)
12. ✅ Create systemd service
13. ✅ Configure firewall
14. ✅ Set packet capture permissions
15. ✅ Create admin user with random password
16. ✅ Start the service

**Installation takes 5-15 minutes** depending on your server speed and internet connection.

#### Step 7: Access Your Application

After installation completes, you'll see a summary with:

**Web Interface URL:**
```
https://YOUR_SERVER_IP:8443
```

**Admin Credentials:**
```bash
cat /etc/protei-monitoring/admin_credentials.txt
```

**Check Service Status:**
```bash
systemctl status protei-monitoring
```

---

## Testing the Deployment

### 1. Verify Service is Running

```bash
# Check service status
systemctl status protei-monitoring

# Should show: "active (running)" in green
```

### 2. Check Logs

```bash
# View live logs
journalctl -u protei-monitoring -f

# View last 50 lines
journalctl -u protei-monitoring -n 50

# Check application log file
tail -f /var/log/protei-monitoring/protei-monitoring.log
```

### 3. Test Web Interface

```bash
# Check if port 8443 is listening
netstat -tlnp | grep 8443

# Test SSL connection
openssl s_client -connect localhost:8443

# Test from remote machine
curl -k https://YOUR_SERVER_IP:8443
```

### 4. Access Web Interface

1. Open browser: `https://YOUR_SERVER_IP:8443`
2. Accept SSL certificate warning (self-signed certificate)
3. Login with admin credentials from `/etc/protei-monitoring/admin_credentials.txt`
4. **Change the password immediately**

### 5. Test Packet Capture

```bash
# Verify capabilities
getcap /opt/protei-monitoring/bin/protei-monitoring

# Should show:
# /opt/protei-monitoring/bin/protei-monitoring = cap_net_raw,cap_net_admin+eip
```

### 6. Test Database Connection

```bash
# Connect to database
sudo -u postgres psql -d protei_monitoring -U protei_user

# List tables (inside psql)
\dt

# Exit
\q
```

### 7. Test Redis Connection

```bash
# Connect to Redis
redis-cli

# Test command
PING
# Should return: PONG

# Exit
exit
```

---

## Source Code Protection

### How Source Code is Protected

1. **Source code stays on your development machine** - it's never deployed to production
2. **Only the compiled binary is deployed** - no Go source code on the server
3. **Binary is stripped** - debugging symbols removed, making reverse engineering harder
4. **Binary is encrypted during transfer** - AES-256-CBC encryption
5. **Optional source encryption** - you can separately encrypt the source code for backup/storage

### Viewing Source Code (On Development Machine Only)

The source code is only accessible on your development machine where you built it.

**Option 1**: Direct access to unencrypted source
```bash
cd /path/to/protei-monitoring
ls -la pkg/
```

**Option 2**: Create encrypted source backup
```bash
cd scripts
./encrypt_source.sh
```

This creates an encrypted archive of your entire source code that only you (with the key) can decrypt.

### Can Others Access the Source Code from the Server?

**NO**. Here's why:

1. **No source files on server** - only the compiled binary exists
2. **Binary is compiled** - not human-readable
3. **Symbols stripped** - function names and debug info removed
4. **Reverse engineering is difficult** - while not impossible, it requires significant expertise and effort
5. **Binary is in a protected directory** - `/opt/protei-monitoring` with restricted permissions

### To View Encrypted Source Code

If you created an encrypted source backup:

```bash
# Decrypt (on your development machine)
cd scripts
./decrypt_source.sh

# Enter your encryption key when prompted
# Source code will be decrypted to specified directory
```

---

## Post-Installation Steps

### 1. Change Default Password

```bash
# Login to web interface
# Go to Settings > Users > Change Password
```

### 2. Replace Self-Signed SSL Certificate

```bash
# Copy your valid certificates to the server
scp your-cert.crt root@YOUR_SERVER:/etc/protei-monitoring/certs/server.crt
scp your-key.key root@YOUR_SERVER:/etc/protei-monitoring/certs/server.key

# Set permissions
ssh root@YOUR_SERVER
chown protei:protei /etc/protei-monitoring/certs/*
chmod 600 /etc/protei-monitoring/certs/server.key

# Restart service
systemctl restart protei-monitoring
```

### 3. Configure Backups

Create a backup script:

```bash
#!/bin/bash
BACKUP_DIR="/backup/protei-monitoring"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup database
sudo -u postgres pg_dump protei_monitoring | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# Backup data directory
tar -czf $BACKUP_DIR/data_$DATE.tar.gz /var/lib/protei-monitoring

# Backup configuration
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /etc/protei-monitoring

# Keep only last 30 days
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
```

Add to crontab:
```bash
0 2 * * * /usr/local/bin/backup-protei.sh
```

### 4. Setup Monitoring

```bash
# Enable metrics endpoint (port 9090)
firewall-cmd --permanent --add-port=9090/tcp
firewall-cmd --reload

# Access metrics
curl http://localhost:9090/metrics
```

### 5. Configure Log Rotation

The application automatically rotates logs, but you can also use logrotate:

```bash
cat > /etc/logrotate.d/protei-monitoring <<'EOF'
/var/log/protei-monitoring/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 protei protei
    sharedscripts
    postrotate
        systemctl reload protei-monitoring
    endscript
}
EOF
```

---

## Troubleshooting

### Issue: Service Won't Start

```bash
# Check status
systemctl status protei-monitoring

# View detailed logs
journalctl -u protei-monitoring -n 100 --no-pager

# Check configuration
/opt/protei-monitoring/bin/protei-monitoring -config /etc/protei-monitoring/config.yaml -check-config

# Common causes:
# - Database not running: systemctl start postgresql
# - Redis not running: systemctl start redis
# - Port already in use: netstat -tlnp | grep 8443
# - Permission issues: chown -R protei:protei /opt/protei-monitoring
```

### Issue: Cannot Access Web Interface

```bash
# Check if service is listening
netstat -tlnp | grep 8443

# Check firewall
firewall-cmd --list-ports  # CentOS/RHEL/Rocky
ufw status                 # Ubuntu/Debian

# Open port if needed
firewall-cmd --permanent --add-port=8443/tcp && firewall-cmd --reload  # CentOS/RHEL
ufw allow 8443/tcp  # Ubuntu/Debian

# Check SELinux (CentOS/RHEL only)
getenforce
# If Enforcing, you may need to configure SELinux policies
```

### Issue: Database Connection Failed

```bash
# Check PostgreSQL is running
systemctl status postgresql

# Check database exists
sudo -u postgres psql -l | grep protei_monitoring

# Check user can connect
sudo -u postgres psql -d protei_monitoring -U protei_user

# Reset password if needed
sudo -u postgres psql <<EOF
ALTER USER protei_user WITH PASSWORD 'new_password';
EOF

# Update config file with new password
vim /etc/protei-monitoring/config.yaml
systemctl restart protei-monitoring
```

### Issue: Packet Capture Not Working

```bash
# Check capabilities
getcap /opt/protei-monitoring/bin/protei-monitoring

# Should show: cap_net_raw,cap_net_admin+eip

# If missing, reapply
setcap 'cap_net_raw,cap_net_admin+eip' /opt/protei-monitoring/bin/protei-monitoring

# Check network interfaces
ip link show

# Test with tcpdump
tcpdump -i any -c 10
```

### Issue: High CPU/Memory Usage

```bash
# Check resource usage
top
htop

# Check specific process
ps aux | grep protei-monitoring

# View process details
systemctl status protei-monitoring

# Adjust limits in systemd service
vim /etc/systemd/system/protei-monitoring.service
# Add under [Service]:
# LimitNOFILE=65536
# LimitNPROC=4096
# MemoryLimit=4G

systemctl daemon-reload
systemctl restart protei-monitoring
```

---

## Service Management Commands

```bash
# Start service
systemctl start protei-monitoring

# Stop service
systemctl stop protei-monitoring

# Restart service
systemctl restart protei-monitoring

# Reload configuration (without restart)
systemctl reload protei-monitoring

# Check status
systemctl status protei-monitoring

# Enable auto-start on boot
systemctl enable protei-monitoring

# Disable auto-start
systemctl disable protei-monitoring

# View logs (live)
journalctl -u protei-monitoring -f

# View logs (last 100 lines)
journalctl -u protei-monitoring -n 100

# View logs (since specific time)
journalctl -u protei-monitoring --since "1 hour ago"

# View logs (specific date)
journalctl -u protei-monitoring --since "2025-01-14"
```

---

## Security Best Practices

1. **Encryption Key Management**
   - Store in password manager
   - Create encrypted backups
   - Never commit to version control
   - Rotate periodically

2. **Access Control**
   - Change default admin password immediately
   - Use strong passwords (16+ characters, mixed case, numbers, symbols)
   - Enable 2FA if available
   - Limit user access based on roles

3. **Network Security**
   - Use firewall rules to restrict access
   - Only allow necessary ports (8443, 9090)
   - Consider VPN access for remote users
   - Use valid SSL certificates in production

4. **System Hardening**
   - Keep OS updated: `apt update && apt upgrade` or `yum update`
   - Disable unnecessary services
   - Configure SELinux/AppArmor
   - Regular security audits

5. **Monitoring & Logging**
   - Monitor application logs regularly
   - Set up alerting for errors/anomalies
   - Enable audit logging
   - Monitor system resources (CPU, RAM, disk)

6. **Backup Strategy**
   - Daily automated backups
   - Test restore procedures
   - Store backups off-site
   - Encrypt backup files

7. **Incident Response**
   - Have a plan for security incidents
   - Document procedures
   - Know who to contact
   - Regular drills

---

## Updating to New Version

When a new version is released:

1. **Build new encrypted package** on development machine
2. **Transfer to server**
3. **Backup current installation**
   ```bash
   cp -a /opt/protei-monitoring /opt/protei-monitoring.backup
   sudo -u postgres pg_dump protei_monitoring > /backup/db_before_upgrade.sql
   ```
4. **Stop service**
   ```bash
   systemctl stop protei-monitoring
   ```
5. **Decrypt and extract new package**
6. **Copy new binary**
   ```bash
   cp bin/protei-monitoring /opt/protei-monitoring/bin/
   ```
7. **Update configuration** if needed
8. **Run database migrations** if any
9. **Start service**
   ```bash
   systemctl start protei-monitoring
   ```
10. **Verify** everything works
11. **Remove backup** after confirming

---

## Product Roadmap

Protei Monitoring continues to evolve with cutting-edge features planned for future releases.

### Planned for Future Releases

The following features are in active development or planned for upcoming versions:

#### 🤖 ML-Based Anomaly Detection (v2.1 - Q2 2025)
- Unsupervised machine learning for automatic anomaly detection
- Baseline learning for normal network behavior patterns
- Real-time anomaly scoring and prediction
- Pattern recognition for recurring issues
- Predictive failure detection (24-hour advance warning)
- Auto-tuning thresholds based on learned patterns

#### 📡 Live Traffic Capture (v2.1 - Q2 2025)
- eBPF-based zero-copy packet capture
- SPAN/port mirroring support
- Multi-interface simultaneous capture
- Kernel bypass with AF_XDP for high performance
- Hardware timestamping for precise latency measurement
- 10 Gbps capture throughput with 0% packet loss

#### 🌊 Kafka Streaming Integration (v2.2 - Q3 2025)
- Real-time event streaming to Apache Kafka
- Per-protocol topic routing
- Avro/Protobuf schema registry support
- Exactly-once delivery semantics
- Integration with big data platforms (Hadoop, Spark)
- Stream to Elasticsearch for advanced search

#### 📊 Grafana Dashboard Templates (v2.1 - Q2 2025)
- 15+ pre-built professional dashboards
- Executive, protocol, procedure, and operational views
- Template variables for dynamic filtering
- Prometheus and PostgreSQL data sources
- Mobile-optimized responsive design
- Pre-configured alerting rules

#### 🚀 6G Protocol Readiness (v2.3+ - 2026+)
- Monitor 3GPP Release 19+ developments
- AI-native protocol support
- Terahertz band extensions
- Quantum-safe cryptography
- Intent-based networking APIs
- Digital twin integration protocols

#### 🌐 Distributed Deployment Support (v2.3 - Q4 2025)
- Multi-node cluster deployment (3-50 nodes)
- Auto-discovery and coordination
- Load balancing across nodes
- Automatic failover on node failure
- State synchronization via Redis/etcd
- Support for 1M+ TPS across cluster
- Geographic distribution for multi-site deployments
- Kubernetes Helm charts and operators

#### 🔒 REST API Rate Limiting (v2.1 - Q2 2025)
- Per-user and per-endpoint rate limits
- Token bucket and sliding window algorithms
- Daily/monthly API quotas
- Burst traffic handling
- IP-based protection
- Configurable limits by user role
- Rate limit metrics and alerting

#### 🏢 Multi-Tenancy Support (v2.2 - Q3 2025)
- Complete data isolation between tenants
- Resource quotas per tenant (CPU, memory, storage)
- White-label UI customization per tenant
- Tenant self-service administration
- Billing integration and usage tracking
- Custom domains per tenant
- Ideal for service providers and MVNOs

#### 📝 Custom Report Builder (v2.2 - Q3 2025)
- Drag-and-drop visual report designer
- No-code report creation
- 20+ chart types (bar, line, pie, heatmap, etc.)
- Scheduled reports with email distribution
- Export to PDF, Excel, CSV, JSON
- 50+ pre-built report templates
- Custom branding and styling

#### 🔐 Advanced LDAP/AD Integration (v2.2 - Q3 2025)
- Multiple LDAP server support
- Full Active Directory schema support
- Group-to-role mapping with nested groups
- Dynamic role assignment
- Connection pooling with failover
- SSL/TLS support (LDAPS)
- Kerberos SSO integration
- User/group synchronization
- Support for Microsoft AD, OpenLDAP, FreeIPA, Azure AD

### Strategic Goals

- **Performance**: 1M TPS per cluster, <10ms latency, 99.99% availability
- **Scalability**: Horizontal scaling to 50+ nodes, multi-region support
- **Intelligence**: 95%+ accuracy in anomaly detection, AI-powered RCA
- **Integration**: 50+ pre-built integrations, universal API
- **User Experience**: <2s page loads, mobile-first design, no-code configuration

For detailed roadmap information, see [ROADMAP.md](ROADMAP.md).

---

## Support

For technical support, questions, or issues:

- **Email**: support@protei-monitoring.com
- **Documentation**: https://docs.protei-monitoring.com
- **GitHub Issues**: (if applicable)

For licensing and sales:
- **Email**: sales@protei-monitoring.com

---

## License

Copyright © 2025 Protei Monitoring. All rights reserved.

This software is proprietary and confidential. Unauthorized copying, distribution, or use is strictly prohibited.

All intellectual property rights are owned by Protei Monitoring.
