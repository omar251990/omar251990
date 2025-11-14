# Protei Monitoring OAM (Operations, Administration & Maintenance) Guide

## Overview

The OAM module provides comprehensive operations, administration, and maintenance capabilities for Protei Monitoring through both CLI scripts and REST API endpoints.

### Key Capabilities

✅ **Application Control**: Start, stop, restart, reload
✅ **Configuration Management**: Read, modify, validate, backup/restore
✅ **Status Monitoring**: Real-time health checks and system metrics
✅ **Log Management**: View and analyze application logs
✅ **Version Control**: Configuration versioning and rollback
✅ **Health Checks**: Comprehensive system health validation
✅ **REST API**: Complete web-based management interface

---

## Table of Contents

1. [Application Control](#application-control)
2. [Configuration Management](#configuration-management)
3. [Status Monitoring](#status-monitoring)
4. [Log Management](#log-management)
5. [Health Checks](#health-checks)
6. [REST API Reference](#rest-api-reference)
7. [CLI Scripts](#cli-scripts)
8. [Best Practices](#best-practices)

---

## Application Control

### CLI Operations

#### Start Application

```bash
# Method 1: Using start script
sudo /usr/protei/Protei_Monitoring/scripts/start

# Method 2: Using systemd
sudo systemctl start protei-monitoring
```

**Output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Starting Protei Monitoring v2.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Application started successfully
PID: 12345
```

#### Stop Application

```bash
# Method 1: Using stop script
sudo /usr/protei/Protei_Monitoring/scripts/stop

# Method 2: Using systemd
sudo systemctl stop protei-monitoring
```

#### Restart Application

```bash
# Method 1: Using restart script
sudo /usr/protei/Protei_Monitoring/scripts/restart

# Method 2: Using systemd
sudo systemctl restart protei-monitoring
```

#### Reload Configuration

```bash
# Reload without restart (sends SIGHUP)
sudo /usr/protei/Protei_Monitoring/scripts/reload
```

#### Check Status

```bash
# Detailed status
sudo /usr/protei/Protei_Monitoring/scripts/status

# Using systemd
sudo systemctl status protei-monitoring
```

**Output:**
```
Protei Monitoring Status:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status:     Running
PID:        12345
Uptime:     2 days 5 hours 30 minutes
Version:    2.0.0
Build Date: 2024-01-15
Git Commit: f1f3a24

System Resources:
  CPU:    15.3%
  Memory: 256 MB
  Disk:   45% used

Network:
  Web Server:  http://0.0.0.0:8080 (OK)
  Database:    PostgreSQL (Connected)
  Redis:       localhost:6379 (Connected)

Health Status: ✅ Healthy
```

### API Operations

#### Start Application

```http
POST /api/v1/oam/app/start
Authorization: Bearer <token>
```

**Response:**
```json
{
  "status": "success",
  "message": "Application started successfully",
  "pid": 12345,
  "start_time": "2024-01-15T10:30:00Z"
}
```

#### Stop Application

```http
POST /api/v1/oam/app/stop
Authorization: Bearer <token>
```

#### Restart Application

```http
POST /api/v1/oam/app/restart
Authorization: Bearer <token>
```

#### Reload Configuration

```http
POST /api/v1/oam/app/reload
Authorization: Bearer <token>
```

#### Get Application Status

```http
GET /api/v1/oam/app/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "status": "running",
  "pid": 12345,
  "uptime_seconds": 185400,
  "start_time": "2024-01-13T05:00:00Z",
  "version": "2.0.0",
  "build_date": "2024-01-15",
  "git_commit": "f1f3a24",
  "restart_count": 0
}
```

---

## Configuration Management

### Configuration Files

| File | Description | Modifiable |
|------|-------------|------------|
| `db.cfg` | Database connection settings | Yes |
| `license.cfg` | License configuration | No (generated) |
| `protocols.cfg` | Protocol enable/disable flags | Yes |
| `system.cfg` | System-wide settings | Yes |
| `trace.cfg` | Packet trace settings | Yes |
| `paths.cfg` | Directory paths | No |
| `security.cfg` | Security settings | Yes |

### View Configuration

#### CLI

```bash
# View entire configuration file
cat /usr/protei/Protei_Monitoring/config/system.cfg

# View specific value
grep "WEB_PORT" /usr/protei/Protei_Monitoring/config/system.cfg
```

#### API

```http
GET /api/v1/oam/config/system.cfg
Authorization: Bearer <token>
```

**Response:**
```json
{
  "filename": "system.cfg",
  "last_modified": "2024-01-15T10:30:00Z",
  "version": 3,
  "content": {
    "WEB_PORT": "8080",
    "LOG_LEVEL": "info",
    "LOG_MAX_SIZE_MB": "100",
    "CDR_FORMAT": "csv",
    "CDR_ROTATION_SIZE_MB": "100",
    "REDIS_ENABLED": "true",
    "REDIS_HOST": "localhost",
    "REDIS_PORT": "6379"
  }
}
```

### Modify Configuration

#### API - Single Value

```http
PUT /api/v1/oam/config/system.cfg/WEB_PORT
Authorization: Bearer <token>
Content-Type: application/json

{
  "value": "8081",
  "changed_by": "admin"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Configuration updated",
  "file": "system.cfg",
  "key": "WEB_PORT",
  "old_value": "8080",
  "new_value": "8081",
  "requires_restart": true
}
```

#### API - Multiple Values

```http
PUT /api/v1/oam/config/system.cfg
Authorization: Bearer <token>
Content-Type: application/json

{
  "values": {
    "WEB_PORT": "8081",
    "LOG_LEVEL": "debug",
    "LOG_MAX_SIZE_MB": "200"
  },
  "changed_by": "admin"
}
```

### Validate Configuration

```http
POST /api/v1/oam/config/validate
Authorization: Bearer <token>
Content-Type: application/json

{
  "file": "system.cfg",
  "values": {
    "WEB_PORT": "99999",
    "LOG_LEVEL": "invalid"
  }
}
```

**Response:**
```json
{
  "valid": false,
  "errors": [
    {
      "file": "system.cfg",
      "key": "WEB_PORT",
      "value": "99999",
      "message": "value must be between 1 and 65535"
    },
    {
      "file": "system.cfg",
      "key": "LOG_LEVEL",
      "value": "invalid",
      "message": "invalid value: invalid (must be one of: [debug info warning error])"
    }
  ]
}
```

### Backup Configuration

```http
POST /api/v1/oam/config/backup
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "Pre-upgrade backup",
  "created_by": "admin"
}
```

**Response:**
```json
{
  "status": "success",
  "backup_time": "2024-01-15T10:30:00Z",
  "files_backed_up": [
    "db.cfg",
    "system.cfg",
    "protocols.cfg",
    "trace.cfg",
    "security.cfg"
  ]
}
```

### Restore Configuration

```http
POST /api/v1/oam/config/restore
Authorization: Bearer <token>
Content-Type: application/json

{
  "backup_time": "2024-01-15T10:30:00Z"
}
```

### View Backup History

```http
GET /api/v1/oam/config/backups
Authorization: Bearer <token>
```

**Response:**
```json
{
  "backups": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "description": "Pre-upgrade backup",
      "created_by": "admin",
      "files": ["db.cfg", "system.cfg", "protocols.cfg"]
    },
    {
      "timestamp": "2024-01-14T09:00:00Z",
      "description": "Daily backup",
      "created_by": "system",
      "files": ["db.cfg", "system.cfg", "protocols.cfg"]
    }
  ]
}
```

---

## Status Monitoring

### System Metrics

```http
GET /api/v1/oam/metrics
Authorization: Bearer <token>
```

**Response:**
```json
{
  "cpu_percent": 15.3,
  "memory_mb": 256,
  "disk_usage": "45%",
  "network_connections": 42,
  "goroutines": 150,
  "threads": 25
}
```

### Process Information

```http
GET /api/v1/oam/process
Authorization: Bearer <token>
```

**Response:**
```json
{
  "pid": 12345,
  "parent_pid": 1,
  "user": "protei",
  "cpu_percent": 15.3,
  "memory_mb": 256,
  "threads": 25,
  "open_files": 120,
  "start_time": "2024-01-13T05:00:00Z",
  "uptime_seconds": 185400
}
```

### Database Status

```http
GET /api/v1/oam/database/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "connected": true,
  "host": "localhost",
  "port": 5432,
  "database": "protei_monitoring",
  "version": "14.5",
  "active_connections": 15,
  "max_connections": 100,
  "database_size": "5.2 GB",
  "response_time_ms": 2
}
```

### Redis Status

```http
GET /api/v1/oam/redis/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "connected": true,
  "host": "localhost",
  "port": 6379,
  "version": "6.2.6",
  "uptime_seconds": 2592000,
  "used_memory_mb": 128,
  "connected_clients": 5,
  "keys": 1250
}
```

---

## Log Management

### View Logs

#### Application Log

```http
GET /api/v1/oam/logs/application?lines=100
Authorization: Bearer <token>
```

**Response:**
```json
{
  "log_type": "application",
  "lines": 100,
  "content": [
    "2024-01-15 10:30:00 [INFO] Application started",
    "2024-01-15 10:30:01 [INFO] Database connected",
    "2024-01-15 10:30:01 [INFO] Redis connected",
    "2024-01-15 10:30:02 [INFO] Web server listening on :8080",
    "..."
  ]
}
```

#### Error Log

```http
GET /api/v1/oam/logs/error?lines=50
Authorization: Bearer <token>
```

#### System Log

```http
GET /api/v1/oam/logs/system?lines=100
Authorization: Bearer <token>
```

#### Access Log

```http
GET /api/v1/oam/logs/access?lines=100
Authorization: Bearer <token>
```

### Download Logs

```http
GET /api/v1/oam/logs/download/application
Authorization: Bearer <token>
```

Returns the full log file for download.

### Log Rotation

```http
POST /api/v1/oam/logs/rotate
Authorization: Bearer <token>
```

Forces immediate log rotation.

---

## Health Checks

### Comprehensive Health Check

```http
GET /api/v1/oam/health
Authorization: Bearer <token>
```

**Response:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "healthy": true,
  "checks": {
    "process": {
      "healthy": true,
      "message": "Process is running"
    },
    "web_server": {
      "healthy": true,
      "message": "Web server is healthy"
    },
    "database": {
      "healthy": true,
      "message": "Database is accessible"
    },
    "disk_space": {
      "healthy": true,
      "message": "Disk usage normal: 45%"
    },
    "redis": {
      "healthy": true,
      "message": "Redis is connected"
    }
  }
}
```

### Liveness Probe

```http
GET /health/live
```

Returns `200 OK` if process is alive.

### Readiness Probe

```http
GET /health/ready
```

Returns `200 OK` if application is ready to serve traffic.

---

## REST API Reference

### Authentication

All OAM endpoints require authentication via Bearer token:

```http
Authorization: Bearer <your_jwt_token>
```

### Base URL

```
https://<server_ip>:8080/api/v1/oam
```

### Endpoints Summary

| Category | Method | Endpoint | Description |
|----------|--------|----------|-------------|
| **App Control** | POST | `/app/start` | Start application |
| | POST | `/app/stop` | Stop application |
| | POST | `/app/restart` | Restart application |
| | POST | `/app/reload` | Reload configuration |
| | GET | `/app/status` | Get application status |
| **Configuration** | GET | `/config/{file}` | Get configuration file |
| | PUT | `/config/{file}/{key}` | Update single value |
| | PUT | `/config/{file}` | Update multiple values |
| | POST | `/config/validate` | Validate configuration |
| | POST | `/config/backup` | Create backup |
| | POST | `/config/restore` | Restore backup |
| | GET | `/config/backups` | List backups |
| **Monitoring** | GET | `/metrics` | Get system metrics |
| | GET | `/process` | Get process info |
| | GET | `/database/status` | Get database status |
| | GET | `/redis/status` | Get Redis status |
| **Logs** | GET | `/logs/{type}` | View logs |
| | GET | `/logs/download/{type}` | Download logs |
| | POST | `/logs/rotate` | Rotate logs |
| **Health** | GET | `/health` | Full health check |

### Error Responses

All endpoints return errors in this format:

```json
{
  "error": {
    "code": "CONFIG_VALIDATION_ERROR",
    "message": "Configuration validation failed",
    "details": {
      "file": "system.cfg",
      "key": "WEB_PORT",
      "value": "99999",
      "validation_error": "value must be between 1 and 65535"
    }
  }
}
```

**HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (validation failed)
- `401` - Unauthorized
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

---

## CLI Scripts

### Control Scripts

Located in `/usr/protei/Protei_Monitoring/scripts/`:

| Script | Description |
|--------|-------------|
| `start` | Start application |
| `stop` | Stop application gracefully |
| `restart` | Restart application |
| `reload` | Reload configuration without restart |
| `status` | Display application status |
| `version` | Display version information |

### Utility Scripts

Located in `/usr/protei/Protei_Monitoring/scripts/utils/`:

| Script | Description |
|--------|-------------|
| `check_db.sh` | Check database connectivity |
| `manage_cdr.sh` | Manage CDR files |
| `encrypt_source.sh` | Encrypt/decrypt source code |
| `generate_license.sh` | Generate license file |
| `backup_config.sh` | Backup configuration files |

---

## Best Practices

### Configuration Changes

1. **Always create a backup before changes:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/oam/config/backup \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"description":"Pre-change backup","created_by":"admin"}'
   ```

2. **Validate before applying:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/oam/config/validate \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"file":"system.cfg","values":{"WEB_PORT":"8081"}}'
   ```

3. **Apply changes:**
   ```bash
   curl -X PUT http://localhost:8080/api/v1/oam/config/system.cfg/WEB_PORT \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"value":"8081","changed_by":"admin"}'
   ```

4. **Reload if possible, restart if necessary:**
   ```bash
   # If change supports hot reload
   curl -X POST http://localhost:8080/api/v1/oam/app/reload \
     -H "Authorization: Bearer $TOKEN"

   # Otherwise, restart
   curl -X POST http://localhost:8080/api/v1/oam/app/restart \
     -H "Authorization: Bearer $TOKEN"
   ```

### Monitoring

1. **Set up regular health checks:**
   ```bash
   */5 * * * * curl -f http://localhost:8080/health/ready || /usr/protei/Protei_Monitoring/scripts/restart
   ```

2. **Monitor system metrics:**
   ```bash
   watch -n 10 'curl -s http://localhost:8080/api/v1/oam/metrics | jq'
   ```

3. **Check logs regularly:**
   ```bash
   tail -f /usr/protei/Protei_Monitoring/logs/application/protei-monitoring.log
   ```

### Maintenance

1. **Daily backup:**
   ```bash
   # Cron job: daily at 2 AM
   0 2 * * * /usr/protei/Protei_Monitoring/scripts/utils/backup_config.sh
   ```

2. **Log rotation:**
   ```bash
   # Rotate logs weekly
   0 0 * * 0 curl -X POST http://localhost:8080/api/v1/oam/logs/rotate
   ```

3. **CDR cleanup:**
   ```bash
   # Clean CDRs older than 90 days
   0 3 * * 0 /usr/protei/Protei_Monitoring/scripts/utils/manage_cdr.sh cleanup 90
   ```

---

## Security Considerations

### API Access Control

- All OAM endpoints require authentication
- Use role-based access control (RBAC):
  - **Admin**: Full access to all OAM operations
  - **Operator**: Read-only access to status and logs
  - **Viewer**: Health checks only

### Audit Logging

All configuration changes are logged:

```sql
SELECT * FROM audit_log
WHERE action LIKE 'config_%'
ORDER BY timestamp DESC
LIMIT 100;
```

### Sensitive Configuration

Files like `db.cfg`, `license.cfg`, and `security.cfg` have restricted permissions (600) and are encrypted at rest.

---

## Troubleshooting

### Issue 1: Cannot Start Application

**Symptoms:** Application fails to start

**Diagnosis:**
```bash
# Check logs
tail -100 /usr/protei/Protei_Monitoring/logs/error/error.log

# Check port conflicts
netstat -tulpn | grep 8080

# Check database connectivity
/usr/protei/Protei_Monitoring/scripts/utils/check_db.sh
```

**Solution:**
- Verify port is not in use
- Check database credentials in `config/db.cfg`
- Ensure sufficient disk space

### Issue 2: Configuration Changes Not Applied

**Symptoms:** Changes don't take effect

**Diagnosis:**
```bash
# Verify file was actually modified
cat /usr/protei/Protei_Monitoring/config/system.cfg | grep WEB_PORT

# Check if reload was successful
curl http://localhost:8080/api/v1/oam/app/status
```

**Solution:**
- Some changes require full restart (not just reload)
- Check if validation passed
- Verify file permissions

### Issue 3: High Memory Usage

**Symptoms:** Application using excessive memory

**Diagnosis:**
```bash
# Check metrics
curl http://localhost:8080/api/v1/oam/metrics

# Check process details
ps aux | grep protei-monitoring
```

**Solution:**
- Adjust correlation engine session timeout
- Reduce log levels
- Increase CDR rotation frequency

---

## Summary

The Protei Monitoring OAM module provides:

✅ **Complete Lifecycle Management** - Start, stop, restart, reload
✅ **Configuration Control** - Modify, validate, backup, restore
✅ **Real-Time Monitoring** - System metrics and health checks
✅ **Log Management** - View, download, rotate logs
✅ **REST API** - Full programmatic access
✅ **CLI Tools** - Script-based management
✅ **Audit Trail** - Track all configuration changes
✅ **Security** - Role-based access control

For more information, see:
- [Installation Guide](INSTALLATION.md)
- [CDR Generation Guide](CDR_GENERATION_GUIDE.md)
- [Correlation Engine Guide](CORRELATION_ENGINE_GUIDE.md)
