# Protei_Bulk Installation Guide

Complete guide for installing and configuring the Protei_Bulk enterprise messaging platform.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Installation Methods](#installation-methods)
3. [Automated Installation](#automated-installation)
4. [Manual Installation](#manual-installation)
5. [Quick Development Setup](#quick-development-setup)
6. [Post-Installation Configuration](#post-installation-configuration)
7. [Verification](#verification)
8. [Troubleshooting](#troubleshooting)
9. [Uninstallation](#uninstallation)

---

## Prerequisites

### Hardware Requirements
- **CPU**: 4+ cores (8+ cores recommended for production)
- **RAM**: 8 GB minimum (16+ GB recommended for production)
- **Disk**: 50 GB minimum (100+ GB recommended for production, SSD preferred)
- **Network**: 1 Gbps network interface

### Software Requirements
- **Operating System**:
  - Ubuntu 18.04+ / Debian 10+
  - CentOS 7+ / RHEL 7+
- **Database**: PostgreSQL 12+ (will be installed automatically)
- **Cache**: Redis 5.0+ (will be installed automatically)
- **Python**: 3.8+ (will be installed automatically)
- **Root Access**: Required for installation

### Network Requirements
- Open ports (for production):
  - 2775 (SMPP)
  - 8080 (HTTP API)
  - 9999 (Management)
  - 5432 (PostgreSQL - internal only)
  - 6379 (Redis - internal only)

---

## Installation Methods

Protei_Bulk provides three installation methods:

| Method | Use Case | Time | Complexity |
|--------|----------|------|------------|
| **Automated** | Production deployment | ~10-15 min | Low |
| **Manual** | Custom configurations | ~30-45 min | High |
| **Quick Dev** | Development/testing | ~5-10 min | Low |

---

## Automated Installation

The automated installation script handles everything for you.

### Step 1: Download/Extract Protei_Bulk

```bash
# If you have a tarball
tar -xzf Protei_Bulk_v1.0.0.tar.gz
cd Protei_Bulk

# If cloning from git
git clone <repository_url>
cd Protei_Bulk
```

### Step 2: Run Installation Script

```bash
sudo ./install.sh
```

The script will:
1. ✓ Check system compatibility
2. ✓ Install system dependencies (PostgreSQL, Redis, Python, etc.)
3. ✓ Create database user: `protei` (password: `elephant`)
4. ✓ Create database: `protei_bulk`
5. ✓ Load database schema (20+ tables)
6. ✓ Load seed data (demo user, templates, etc.)
7. ✓ Set up Python virtual environment
8. ✓ Install Python dependencies (~50 packages)
9. ✓ Configure application settings
10. ✓ Create application user: `protei`
11. ✓ Set up systemd service
12. ✓ Verify installation

### Installation Process

```
╔════════════════════════════════════════════════════════════════╗
║          Protei_Bulk Installation                              ║
╚════════════════════════════════════════════════════════════════╝

This script will install and configure Protei_Bulk with the following:

  • PostgreSQL database server
  • Redis server
  • Python 3.8 and dependencies
  • Database: protei_bulk
  • Database User: protei
  • Installation Directory: /opt/Protei_Bulk

Do you want to continue? [y/N] y

═══════════════════════════════════════════════════════════════
  Installing System Dependencies
═══════════════════════════════════════════════════════════════

[✓] System dependencies installed

═══════════════════════════════════════════════════════════════
  Setting Up PostgreSQL
═══════════════════════════════════════════════════════════════

[✓] PostgreSQL service started
[✓] Database user created: protei
[✓] Database created: protei_bulk
[✓] PostgreSQL configured

═══════════════════════════════════════════════════════════════
  Setting Up Database Schema
═══════════════════════════════════════════════════════════════

[✓] Database schema loaded successfully
[i] Created 23 database tables

... (continues)

╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║          ✓ Installation Completed Successfully!               ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
```

### Step 3: Load Seed Data (Optional)

```bash
sudo -u protei ./scripts/utils/load_seed_data.sh
```

This creates:
- Default admin user (username: `admin`, password: `Admin@123`)
- Sample message templates
- Demo SMSC connection
- Sample contact lists
- System configuration

### Step 4: Start the Service

```bash
sudo systemctl start protei_bulk
sudo systemctl status protei_bulk
```

Or manually:

```bash
sudo -u protei ./scripts/start
```

---

## Manual Installation

For custom installations or when you need more control.

### Step 1: Install System Dependencies

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    python3 \
    python3-pip \
    python3-dev \
    python3-venv \
    postgresql \
    postgresql-contrib \
    redis-server \
    libpq-dev \
    git \
    curl
```

#### CentOS/RHEL

```bash
sudo yum update -y
sudo yum install -y \
    gcc \
    python3 \
    python3-pip \
    python3-devel \
    postgresql-server \
    postgresql-contrib \
    redis \
    postgresql-devel \
    git \
    curl
```

### Step 2: Set Up PostgreSQL

```bash
# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create user and database
sudo -u postgres psql <<EOF
CREATE USER protei WITH PASSWORD 'elephant';
ALTER USER protei WITH SUPERUSER;
CREATE DATABASE protei_bulk OWNER protei;
\q
EOF
```

### Step 3: Configure PostgreSQL Authentication

Edit `/etc/postgresql/[version]/main/pg_hba.conf` (Ubuntu) or `/var/lib/pgsql/data/pg_hba.conf` (CentOS):

Add these lines:

```
host    protei_bulk    protei    127.0.0.1/32    md5
host    protei_bulk    protei    ::1/128         md5
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

### Step 4: Load Database Schema

```bash
export PGPASSWORD='elephant'
psql -h localhost -U protei -d protei_bulk -f database/schema.sql
unset PGPASSWORD
```

### Step 5: Load Seed Data

```bash
export PGPASSWORD='elephant'
psql -h localhost -U protei -d protei_bulk -f database/seed_data.sql
unset PGPASSWORD
```

### Step 6: Set Up Python Environment

```bash
# Create virtual environment
python3 -m venv venv

# Activate
source venv/bin/activate

# Upgrade pip
pip install --upgrade pip

# Install dependencies
pip install -r requirements.txt

# Deactivate
deactivate
```

### Step 7: Configure Application

Edit `config/db.conf`:

```ini
[PostgreSQL]
host = localhost
port = 5432
database = protei_bulk
username = protei
password = elephant
```

### Step 8: Create Application User

```bash
sudo useradd -r -s /bin/bash -d /opt/Protei_Bulk protei
sudo chown -R protei:protei /opt/Protei_Bulk
```

### Step 9: Set Up Systemd Service (Optional)

Create `/etc/systemd/system/protei_bulk.service`:

```ini
[Unit]
Description=Protei Bulk Messaging Platform
After=network.target postgresql.service redis.service

[Service]
Type=forking
User=protei
Group=protei
WorkingDirectory=/opt/Protei_Bulk
ExecStart=/opt/Protei_Bulk/scripts/start
ExecStop=/opt/Protei_Bulk/scripts/stop
PIDFile=/opt/Protei_Bulk/tmp/protei_bulk.pid
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable protei_bulk
sudo systemctl start protei_bulk
```

---

## Quick Development Setup

For developers who want to quickly set up a development environment.

### Prerequisites

- PostgreSQL already installed and running
- Python 3.8+ already installed
- Redis already installed (optional)

### Run Quick Setup

```bash
./quick_dev_setup.sh
```

This script:
- Creates database and user
- Loads schema and seed data
- Sets up Python virtual environment
- Installs dependencies
- Configures application

**Total time**: ~5-10 minutes

### Manual Dev Setup (Alternative)

If you prefer manual control:

```bash
# 1. Create database
sudo -u postgres createuser -P protei  # Password: elephant
sudo -u postgres createdb -O protei protei_bulk

# 2. Load schema
PGPASSWORD=elephant psql -h localhost -U protei -d protei_bulk < database/schema.sql

# 3. Load seed data
PGPASSWORD=elephant psql -h localhost -U protei -d protei_bulk < database/seed_data.sql

# 4. Create virtual environment
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# 5. Run application
./bin/Protei_Bulk
```

---

## Post-Installation Configuration

### 1. Change Default Password

Default credentials:
- Username: `admin`
- Password: `Admin@123`

**Change immediately on first login!**

### 2. Update License Key

Replace the demo license with your production license:

```bash
sudo nano config/license.key
```

Verify license:

```bash
./scripts/utils/check_license.sh
```

### 3. Configure SMSC Connections

Edit `config/protocol.conf`:

```ini
[SMPP]
enabled = true
bind_address = 0.0.0.0
bind_port = 2775
system_id = YOUR_SYSTEM_ID
```

### 4. Configure Email/SMS Alerts

Edit `config/security.conf` for alert notifications.

### 5. Set Up Firewall

```bash
# Allow SMPP
sudo ufw allow 2775/tcp

# Allow HTTP API (restrict to your network)
sudo ufw allow from 192.168.1.0/24 to any port 8080

# Allow HTTPS (if configured)
sudo ufw allow 8443/tcp
```

### 6. Configure Log Rotation

The system includes automatic log rotation. Adjust settings in `config/log.conf`:

```ini
[Rotation]
enable_rotation = true
max_size_mb = 100
max_backups = 10
retention_days = 90
```

---

## Verification

### Check System Status

```bash
# Using systemd
sudo systemctl status protei_bulk

# Using scripts
./scripts/status

# Check logs
tail -f logs/system.log
tail -f logs/startup.log
```

### Verify Database Connection

```bash
export PGPASSWORD='elephant'
psql -h localhost -U protei -d protei_bulk -c "SELECT COUNT(*) FROM users;"
unset PGPASSWORD
```

Expected output: At least 1 user (admin)

### Verify Services

```bash
# PostgreSQL
sudo systemctl status postgresql

# Redis
sudo systemctl status redis-server  # or redis

# Protei_Bulk
sudo systemctl status protei_bulk
```

### Test API (if configured)

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 123
}
```

### Check Database Tables

```bash
export PGPASSWORD='elephant'
psql -h localhost -U protei -d protei_bulk -c "\dt"
unset PGPASSWORD
```

Should show 20+ tables.

---

## Troubleshooting

### PostgreSQL Connection Failed

**Problem**: Cannot connect to database

**Solutions**:
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check if listening
sudo netstat -tlnp | grep 5432

# Check pg_hba.conf
sudo nano /etc/postgresql/[version]/main/pg_hba.conf

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### Permission Denied Errors

**Problem**: Scripts cannot be executed

**Solutions**:
```bash
# Make scripts executable
chmod +x install.sh
chmod +x scripts/*
chmod +x scripts/utils/*

# Check ownership
ls -la | grep scripts
sudo chown -R protei:protei /opt/Protei_Bulk
```

### Python Dependencies Failed

**Problem**: pip install errors

**Solutions**:
```bash
# Upgrade pip
pip install --upgrade pip

# Install system packages
sudo apt-get install -y python3-dev libpq-dev  # Ubuntu
sudo yum install -y python3-devel postgresql-devel  # CentOS

# Retry installation
pip install -r requirements.txt
```

### Service Won't Start

**Problem**: `systemctl start protei_bulk` fails

**Solutions**:
```bash
# Check logs
journalctl -u protei_bulk -n 50

# Check PID file
cat tmp/protei_bulk.pid
ps aux | grep Protei_Bulk

# Remove stale PID
rm -f tmp/protei_bulk.pid

# Try manual start
sudo -u protei ./scripts/start
```

### Port Already in Use

**Problem**: Port 2775, 8080, etc. already occupied

**Solutions**:
```bash
# Find process using port
sudo lsof -i :2775
sudo lsof -i :8080

# Kill process or change port in config/protocol.conf
```

### Schema Load Errors

**Problem**: Database schema fails to load

**Solutions**:
```bash
# Drop and recreate database
sudo -u postgres psql -c "DROP DATABASE protei_bulk;"
sudo -u postgres psql -c "CREATE DATABASE protei_bulk OWNER protei;"

# Reload schema
export PGPASSWORD='elephant'
psql -h localhost -U protei -d protei_bulk < database/schema.sql
```

---

## Uninstallation

### Automated Uninstall

```bash
sudo ./uninstall.sh
```

**Warning**: This removes:
- Protei_Bulk application
- Database and user
- System service
- Application user

### Manual Uninstall

```bash
# Stop service
sudo systemctl stop protei_bulk
sudo systemctl disable protei_bulk

# Remove service file
sudo rm /etc/systemd/system/protei_bulk.service
sudo systemctl daemon-reload

# Drop database
sudo -u postgres psql -c "DROP DATABASE protei_bulk;"
sudo -u postgres psql -c "DROP USER protei;"

# Remove application
sudo rm -rf /opt/Protei_Bulk

# Remove user
sudo userdel protei
```

---

## Support

### Documentation
- Installation Guide: `document/Installation_Guide.docx`
- Deployment Manual: `document/Deployment_Manual.docx`
- API Reference: `document/API_Reference.docx`
- System Design: `document/System_Design_Document.docx`

### Contact
- Email: support@protei.com
- Website: https://www.protei.com
- Documentation: https://docs.protei.com/protei-bulk/

### Logs
- Installation: `installation.log`
- Application: `logs/system.log`
- Errors: `logs/error.log`
- Startup: `logs/startup.log`

---

**Last Updated**: 2025-01-16
**Version**: 1.0.0
