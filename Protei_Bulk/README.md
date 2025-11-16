# Protei_Bulk

**High-Performance Bulk Messaging Platform**

Protei_Bulk is an enterprise-grade bulk messaging and protocol handling system designed for telecommunications operators and service providers.

## Features

- ✓ **Multi-Protocol Support**: SMPP 3.4, HTTP REST API
- ✓ **High Performance**: 10,000+ messages per second
- ✓ **Campaign Management**: Create and manage bulk messaging campaigns
- ✓ **Real-time CDR**: Comprehensive Call Detail Record generation
- ✓ **Scalable Architecture**: Horizontal scaling support
- ✓ **High Availability**: Active-standby and active-active configurations
- ✓ **Security**: Role-based access control, encryption, audit logging
- ✓ **Monitoring**: Built-in performance metrics and health checks

## Quick Start

### Installation

1. Extract the package:
   ```bash
   tar -xzf Protei_Bulk_v1.0.0.tar.gz
   cd Protei_Bulk
   ```

2. Configure the application:
   ```bash
   # Edit configuration files in config/
   vi config/app.conf
   vi config/db.conf
   vi config/protocol.conf
   ```

3. Install license:
   ```bash
   cp your_license.key config/license.key
   scripts/utils/check_license.sh
   ```

4. Start the service:
   ```bash
   scripts/start
   ```

### Usage

Check service status:
```bash
scripts/status
```

Stop the service:
```bash
scripts/stop
```

Restart the service:
```bash
scripts/restart
```

Reload configuration:
```bash
scripts/reload
```

View version information:
```bash
scripts/version
```

## Directory Structure

```
Protei_Bulk/
├── bin/                  # Application binaries
├── config/               # Configuration files
├── lib/                  # Libraries and dependencies
├── cdr/                  # Call Detail Records
│   ├── smpp/
│   ├── http/
│   ├── internal/
│   └── archive/
├── logs/                 # Application logs
├── scripts/              # Management scripts
│   ├── start
│   ├── stop
│   ├── restart
│   ├── reload
│   ├── status
│   ├── version
│   └── utils/
├── tmp/                  # Temporary files
└── document/             # Documentation
```

## Configuration

### Main Configuration Files

- **app.conf**: Core application settings
- **db.conf**: Database connection parameters
- **log.conf**: Logging configuration
- **protocol.conf**: Protocol-specific settings (SMPP, HTTP)
- **network.conf**: Network interfaces and ports
- **security.conf**: Authentication and authorization
- **license.key**: Application license

## API Endpoints

### HTTP REST API

Base URL: `http://your-server:8080/api/v1`

- `GET /health` - Health check
- `POST /messages` - Submit single message
- `POST /messages/bulk` - Submit bulk messages
- `GET /messages/{id}` - Get message status
- `POST /campaigns` - Create campaign
- `GET /campaigns/{id}` - Get campaign status
- `GET /statistics` - Get system statistics

### SMPP Protocol

- **Host**: your-server
- **Port**: 2775
- **Version**: SMPP 3.4
- **Supported Operations**: bind_transmitter, bind_receiver, bind_transceiver, submit_sm, deliver_sm

## System Requirements

### Hardware
- CPU: 4+ cores (8+ recommended)
- RAM: 8 GB minimum (16+ GB recommended)
- Disk: 50 GB minimum (SSD recommended)
- Network: 1 Gbps interface

### Software
- OS: Linux (Ubuntu 20.04+, CentOS 7+)
- Python: 3.8+
- Database: PostgreSQL 12+ or MySQL 8+
- Redis: 5.0+

## Documentation

Complete documentation is available in the `document/` directory:

- **Installation_Guide.docx**: Step-by-step installation instructions
- **Deployment_Manual.docx**: Production deployment guide
- **API_Reference.docx**: Complete API documentation
- **Web_User_Manual.docx**: Web interface user guide
- **System_Design_Document.docx**: Technical architecture
- **Change_Log.docx**: Version history and updates
- **License_Notes.docx**: Licensing information

## Utilities

### Database Backup
```bash
scripts/utils/backup_db.sh
```

### Log Rotation
```bash
scripts/utils/rotate_logs.sh
```

### License Check
```bash
scripts/utils/check_license.sh
```

### Cleanup Temporary Files
```bash
scripts/utils/cleanup_tmp.sh
```

## Monitoring

### View System Logs
```bash
tail -f logs/system.log
```

### View Error Logs
```bash
tail -f logs/error.log
```

### View Startup Logs
```bash
tail -f logs/startup.log
```

### Monitor CDRs
```bash
tail -f cdr/smpp/*.cdr
```

## Support

### Contact Information
- **Email**: support@protei.com
- **Website**: https://www.protei.com
- **Documentation**: https://docs.protei.com/protei-bulk/

### License
- For licensing information, see `document/License_Notes.docx`
- To obtain a license, contact: sales@protei.com

## Version

**Version**: 1.0.0
**Build**: 001
**Release Date**: 2025-01-16

---

© 2025 Protei Corporation. All rights reserved.
