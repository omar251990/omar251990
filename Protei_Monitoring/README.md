# Protei Monitoring v2.0

**Professional Telecom Signaling Monitoring & Analysis System**

![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)
![License](https://img.shields.io/badge/license-Commercial-green.svg)
![Status](https://img.shields.io/badge/status-Production-brightgreen.svg)

---

## 📋 Overview

Protei Monitoring is an enterprise-grade telecom signaling monitoring system that provides real-time capture, decode, analysis, and visualization of telecom protocols across 2G, 3G, 4G, and 5G networks.

### Key Features

#### 🔌 Protocol Support (10 Protocols)
- **2G/3G**: MAP, CAP, INAP
- **4G/5G Signaling**: Diameter, HTTP/2 (SBI)
- **Data Plane**: GTP, PFCP
- **RAN**: NGAP (5G), S1AP (4G)
- **Mobile**: NAS (4G/5G)

#### 🤖 AI & Intelligence
- **Knowledge Base**: 18 built-in 3GPP standards and IETF RFCs
- **AI Analysis Engine**: 7 intelligent detection rules for automatic issue identification
- **Flow Reconstruction**: 5 standard procedure templates with deviation detection
- **Subscriber Correlation**: Multi-identifier tracking (IMSI/MSISDN/IMEI/TEID/SEID)

#### 🌐 Web Interface
- Real-time dashboard with live statistics
- Advanced search and filtering
- Interactive ladder diagram visualization
- AI-powered issue detection and recommendations
- Comprehensive user management with RBAC

#### 🔒 Enterprise Security
- MAC address-based license binding
- Source code encryption (AES-256-CBC)
- LDAP/Active Directory integration
- Comprehensive audit logging
- Role-based access control (Admin/Operator/Viewer)

#### 📊 Analytics & Reporting
- Real-time KPI monitoring
- Custom report builder
- Automated CDR generation (per protocol)
- Data export (CSV, JSON, XML, PCAP, PDF)

---

## 🚀 Quick Start

### Prerequisites

- **OS**: RHEL/CentOS 8+, Ubuntu 20.04+, Debian 11+
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Network**: Access to SPAN/mirror port or TAP device
- **License**: Valid license file from Protei

### Installation (5 Minutes)

```bash
# 1. Extract package
cd /home/user/omar251990
sudo tar -xzf Protei_Monitoring-2.0.0.tar.gz -C /usr/protei/

# 2. Navigate to directory
cd /usr/protei/Protei_Monitoring

# 3. Configure license
sudo nano config/license.cfg
# Update: LICENSE_MAC and LICENSE_EXPIRY

# 4. Configure database
sudo nano config/db.cfg
# Update: DB_HOST, DB_USER, DB_PASSWORD

# 5. Start application
sudo chmod +x scripts/*
sudo scripts/start
```

### Access Web Interface

```
http://<server_ip>:8080

Default Login:
  Username: admin
  Password: admin (change immediately!)
```

**See [Quick Start Guide](document/QUICK_START.md) for detailed instructions.**

---

## 📁 Directory Structure

```
/usr/protei/Protei_Monitoring/
│
├── bin/                    # Application binaries and source code
│   ├── cmd/               # Main application
│   ├── pkg/               # All packages (decoders, analysis, web, etc.)
│   └── internal/          # Internal packages
│
├── config/                # Configuration files
│   ├── license.cfg       # License configuration
│   ├── db.cfg            # Database settings
│   ├── protocols.cfg     # Protocol enable/disable
│   ├── system.cfg        # System parameters
│   ├── trace.cfg         # Logging configuration
│   ├── paths.cfg         # File paths
│   └── security.cfg      # Security settings
│
├── lib/                   # External libraries and dependencies
│
├── cdr/                   # CDR output files (per protocol)
│   ├── MAP/
│   ├── CAP/
│   ├── Diameter/
│   ├── GTP/
│   └── ... (10 protocols + combined)
│
├── logs/                  # Application logs
│   ├── application/      # General application logs
│   ├── system/           # System-level logs
│   ├── debug/            # Debug logs
│   ├── error/            # Error logs
│   └── access/           # Web access logs
│
├── scripts/               # Control scripts
│   ├── start*            # Start application
│   ├── stop*             # Stop application
│   ├── restart*          # Restart application
│   ├── reload*           # Reload configuration (no downtime)
│   ├── status*           # Check application status
│   ├── version*          # Show version information
│   └── utils/            # Utility scripts
│
├── tmp/                   # Temporary files
│
├── document/              # Complete documentation
│   ├── README.md         # Documentation index
│   ├── INSTALLATION_GUIDE.md
│   ├── QUICK_START.md
│   ├── USER_MANUAL.md
│   ├── WEB_INTERFACE_GUIDE.md
│   ├── API_REFERENCE.md
│   └── ... (30+ documents)
│
└── README.md             # This file
```

---

## 🎯 Core Capabilities

### 1. Protocol Decoding

Full decode support for 10 telecom protocols:

| Protocol | Standards | Use Case | Status |
|----------|-----------|----------|--------|
| **MAP** | 3GPP TS 29.002 | 2G/3G location, SMS, supplementary services | ✅ Full |
| **CAP** | 3GPP TS 29.078 | CAMEL prepaid, IN services | ✅ Full |
| **INAP** | ITU-T Q.1218 | Intelligent network services | ✅ Full |
| **Diameter** | RFC 6733, 3GPP TS 29.272/273 | 4G/5G authentication, mobility | ✅ Full |
| **GTP** | 3GPP TS 29.274, 29.281 | 4G/5G data tunneling | ✅ Full |
| **PFCP** | 3GPP TS 29.244 | 5G UPF control | ✅ Full |
| **HTTP/2** | 3GPP TS 29.500 | 5G Service-Based Interface | ✅ Full |
| **NGAP** | 3GPP TS 38.413 | 5G RAN signaling | ✅ Full |
| **S1AP** | 3GPP TS 36.413 | 4G RAN signaling | ✅ Full |
| **NAS** | 3GPP TS 24.301, 24.501 | 4G/5G mobile signaling | ✅ Full |

### 2. AI-Powered Analysis

**Intelligent Detection Rules:**
1. Repeated failures (same error, same subscriber)
2. Timeout patterns (slow responses across sessions)
3. Authentication failures (security issues)
4. Roaming anomalies (unexpected location changes)
5. High error rates (protocol-specific thresholds)
6. Missing procedures (incomplete flows)
7. Parameter anomalies (out-of-range values)

**Automatic Root Cause Analysis:**
- Pattern recognition across sessions
- Historical trend correlation
- Knowledge base lookup
- Actionable recommendations

### 3. Flow Reconstruction

**Standard Procedure Templates:**
1. **4G Attach** - Initial attach with authentication
2. **5G Registration** - Initial/periodic/mobility registration
3. **PDU Session Establishment** - Data session setup
4. **GTP Tunnel Creation** - Bearer establishment
5. **MAP Location Update** - Mobility management

**Deviation Detection:**
- Missing messages
- Incorrect message order
- Timing anomalies
- Extra/unexpected messages
- Completeness scoring

### 4. Subscriber Intelligence

**Multi-Identifier Tracking:**
- IMSI (International Mobile Subscriber Identity)
- MSISDN (Phone number)
- IMEI (Device identifier)
- TEID (GTP Tunnel Endpoint ID)
- SEID (PFCP Session Endpoint ID)

**Timeline View:**
- All subscriber events chronologically
- Visual markers for different event types
- Location history
- Active sessions
- Issue correlation

---

## 📖 Documentation

Complete documentation available in `document/` directory:

### Getting Started
- [Quick Start Guide](document/QUICK_START.md) - 5-minute setup
- [Installation Guide](document/INSTALLATION_GUIDE.md) - Detailed installation
- [System Requirements](document/SYSTEM_REQUIREMENTS.md) - Hardware/software needs

### User Guides
- [User Manual](document/USER_MANUAL.md) - Complete user guide
- [Web Interface Guide](document/WEB_INTERFACE_GUIDE.md) - Web UI navigation
- [API Reference](document/API_REFERENCE.md) - REST API documentation

### Administrator Guides
- [Administrator Manual](document/ADMIN_MANUAL.md) - System administration
- [Configuration Guide](document/CONFIGURATION_GUIDE.md) - All configuration options
- [Deployment Guide](document/DEPLOYMENT_GUIDE.md) - Production deployment

### Technical Documentation
- [Architecture Overview](document/ARCHITECTURE.md) - System design
- [Protocol Support](document/PROTOCOL_SUPPORT.md) - Supported protocols
- [Application Structure](document/APPLICATION_STRUCTURE.md) - Code structure

### Reference
- [Roadmap](document/ROADMAP.md) - Future features (v2.1-v2.3)
- [Verification & Testing](document/VERIFICATION_TESTING_GUIDE.md) - Testing guide
- [Deployment README](document/DEPLOYMENT_README.md) - Deployment details

---

## 🔧 Common Operations

### Start/Stop/Restart

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

### View Logs

```bash
# Application log
tail -f logs/application/protei-monitoring.log

# Error log
tail -f logs/error/error.log

# Access log (web)
tail -f logs/access/access.log
```

### Export CDRs

```bash
# CDRs are automatically saved to:
cdr/{PROTOCOL}/YYYYMMDD_*.cdr

# Example:
cdr/MAP/20251114_location_updates.cdr
cdr/Diameter/20251114_authentication.cdr
cdr/GTP/20251114_data_sessions.cdr
```

---

## 🌐 API Access

REST API available at: `http://<server>:8080/api/`

### Quick API Examples

```bash
# Get system health
curl http://localhost:8080/health

# List all protocols
curl http://localhost:8080/api/protocols

# Get recent sessions
curl http://localhost:8080/api/sessions?limit=10

# Search by IMSI
curl "http://localhost:8080/api/subscribers?imsi=123456789012345"

# Get AI-detected issues
curl http://localhost:8080/api/analysis/issues

# List knowledge base standards
curl http://localhost:8080/api/knowledge/standards

# Get flow reconstruction templates
curl http://localhost:8080/api/flows/templates
```

**See [API Reference](document/API_REFERENCE.md) for complete API documentation.**

---

## 🔐 Security

### License Protection
- **MAC Address Binding**: License tied to server hardware
- **Expiry Validation**: Automatic license expiry checks
- **Feature Licensing**: Protocol/feature enablement based on license

### Application Security
- **Source Encryption**: AES-256-CBC encrypted binaries
- **Authentication**: JWT-based token authentication
- **Authorization**: Role-based access control (RBAC)
- **Audit Logging**: Complete audit trail of all actions

### Network Security
- **HTTPS Support**: TLS 1.2/1.3 support for web interface
- **LDAP/AD Integration**: Enterprise authentication
- **Session Management**: Secure session handling with timeout
- **Password Policies**: Configurable complexity requirements

---

## 📊 System Requirements

### Minimum Requirements
- **CPU**: 4 cores @ 2.4 GHz
- **RAM**: 8 GB
- **Disk**: 100 GB SSD
- **Network**: 1 Gbps NIC

### Recommended (Production)
- **CPU**: 16 cores @ 3.0 GHz+
- **RAM**: 32 GB+
- **Disk**: 500 GB NVMe SSD (RAID 10)
- **Network**: 10 Gbps NIC with SPAN/mirror access

### Software Dependencies
- PostgreSQL 14+ (database)
- Redis 6+ (caching)
- libpcap 1.10+ (packet capture)
- Go 1.21+ (if building from source)

---

## 🗺️ Roadmap

### Version 2.1 (Q2 2025)
- ML-based anomaly detection
- Live traffic capture (eBPF/SPAN)
- Grafana dashboard templates
- API rate limiting

### Version 2.2 (Q3 2025)
- Kafka streaming integration
- Multi-tenancy support
- Custom report builder
- Advanced LDAP/AD integration

### Version 2.3 (Q4 2025)
- Distributed deployment support
- High availability (HA) clustering
- Geographic redundancy
- Enhanced performance

**See [ROADMAP.md](document/ROADMAP.md) for detailed feature planning.**

---

## 📞 Support

### Documentation
- **Local Docs**: `document/` directory (30+ documents)
- **Online Docs**: https://docs.protei.com
- **API Reference**: [API_REFERENCE.md](document/API_REFERENCE.md)

### Contact
- **Email**: support@protei.com
- **Emergency**: +1-XXX-XXX-XXXX (24/7)
- **Sales**: sales@protei.com
- **Issues**: https://github.com/protei/monitoring/issues

### Training
- **Online Tutorials**: Available in web interface
- **On-site Training**: Contact sales
- **Certification Program**: Available for operators

---

## 📄 License

**Commercial License** - Protei © 2025

This is commercial software. Unauthorized copying, distribution, or modification is prohibited. See license file for details.

---

## 🏆 About Protei

Protei specializes in telecom network monitoring, testing, and analytics solutions. With over 20 years of experience in the telecom industry, we provide professional-grade tools for network operators, equipment vendors, and service providers.

**Our Solutions:**
- Signaling monitoring (2G/3G/4G/5G)
- Protocol testing and validation
- Network optimization
- Security monitoring
- Fraud detection

**Website**: https://www.protei.com

---

## Version Information

- **Version**: 2.0.0
- **Release Date**: November 2025
- **Build**: Production
- **Go Version**: 1.21+
- **License**: Commercial

---

**© 2025 Protei. All rights reserved.**
