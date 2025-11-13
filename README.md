# Protei_Monitoring v2.0

🌐 **Full Telecom-Grade Multi-Protocol Monitoring & Analysis Platform**

A comprehensive monitoring solution for 2G/3G/4G/5G networks with deep protocol decoding, intelligent correlation, KPI analytics, real-time visualization, and enterprise-grade security.

## 📋 Overview

Protei_Monitoring is a carrier-grade platform capable of:

- **Multi-Protocol Decoding**: MAP, CAP, INAP, Diameter, GTP-C, PFCP, HTTP, NGAP, S1AP, NAS
- **Intelligent Correlation**: Automatic session tracking with unique Transaction IDs (TID)
- **KPI Analytics**: Success rates, latency metrics, failure analysis, cause distribution
- **Roaming Intelligence**: Inbound/outbound roamer tracking with cell-level heatmaps
- **Real-time Visualization**: Ladder diagrams, network flow graphs, interactive dashboard
- **Vendor Support**: Ericsson, Huawei, ZTE, Nokia equipment with extensible dictionaries
- **Production-Ready**: Self-contained binary, graceful shutdown, health monitoring, automatic log rotation
- **PCAP Processing**: File-based capture with automatic directory monitoring
- **Database Integration**: PostgreSQL support with Liquibase-style migrations (optional)
- **Enterprise Security**: JWT authentication, RBAC, MAC-based license validation
- **Vendor Dictionaries**: YAML-based extensible AVP/IE definitions for all major vendors

## 🚀 Quick Start

### Build and Install

```bash
# Build the application
make all

# Generate a license (required for v2.0)
./bin/generate_license \
  --customer "YourCompany" \
  --expiry "2030-12-31" \
  --mac $(ip link show | grep ether | head -1 | awk '{print $2}') \
  --2g --3g --4g --5g \
  --map --cap --inap --diameter --gtp --http \
  --max-subscribers 5000000 \
  --max-tps 5000 \
  --output configs/license.json

# Install to /usr/protei/Protei_Monitoring
sudo make install

# Start the service
sudo /usr/protei/Protei_Monitoring/scripts/start.sh
```

### Access Dashboard

Open your browser to: `http://localhost:8080`

**Default credentials**: admin / admin (change on first login)

## ✨ Key Features

### Protocol Support

| Protocol | Version | Interface | Description |
|----------|---------|-----------|-------------|
| MAP | 2, 3 | SS7 | Mobile Application Part (HLR/VLR) |
| CAP | 1-4 | SS7 | CAMEL Application Part (IN) |
| INAP | 1-3 | SS7 | Intelligent Network Application Part |
| Diameter | All | S6a/S6d/Gx/Gy/Gz/S8/S9 | Authentication, policy, charging |
| GTP-C | v1, v2 | S5/S8/S11 | Bearer management |
| PFCP | v1 | N4/Sxa/Sxb | User plane control |
| HTTP | 1.1, 2.0 | 5G SBA | Service-based architecture |
| NGAP | - | N2 | 5G control plane |
| S1AP | - | S1 | 4G control plane |
| NAS | 4G, 5G | Air interface | Non-access stratum |

### Analytics Capabilities

- **Procedure KPIs**:
  - 4G Attach / 5G Registration
  - PDN/PDU Session Establishment
  - Handover (X2/Xn/S1/N2)
  - Location Update / Tracking Area Update
  - Authentication / Service Request

- **Performance Metrics**:
  - Success/Failure rates
  - Latency (Average, P95, P99)
  - Cause code distribution
  - Message throughput

- **Roaming Analytics**:
  - Inbound/Outbound roamer counts
  - PLMN-based success rates
  - Cell-level heatmaps
  - APN usage patterns

### Visualization

- **Ladder Diagrams**: Interactive SVG-based message flow visualization
- **Network Topology**: Automatic node identification and path tracking
- **Real-time Dashboard**: Live KPI updates, session counts, alerts
- **Heatmaps**: Geographic distribution of roaming activity

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Protei_Monitoring                        │
├─────────────────────────────────────────────────────────────┤
│  Input Layer        │  PCAP Files / Live Capture            │
├─────────────────────────────────────────────────────────────┤
│  Decoder Layer      │  MAP │ Diameter │ GTP │ HTTP │ NGAP  │
├─────────────────────────────────────────────────────────────┤
│  Correlation        │  TID Generation & Session Tracking    │
├─────────────────────────────────────────────────────────────┤
│  Analytics          │  KPI Calculation & Failure Analysis   │
├─────────────────────────────────────────────────────────────┤
│  Storage            │  Events (JSONL) │ CDR (CSV) │ Logs   │
├─────────────────────────────────────────────────────────────┤
│  Visualization      │  Ladder Diagrams │ Heatmaps │ Charts │
├─────────────────────────────────────────────────────────────┤
│  Output             │  Web Dashboard │ REST API │ Metrics  │
└─────────────────────────────────────────────────────────────┘
```

## 📦 Installation

See [INSTALLATION.md](docs/INSTALLATION.md) for detailed instructions.

### System Requirements

- **OS**: RHEL 9.6 / Ubuntu 22.04 or compatible
- **CPU**: 4+ cores (8+ recommended)
- **RAM**: 8GB minimum (16GB+ recommended)
- **Disk**: 100GB+ for logs and CDR storage
- **Go**: 1.21+ (for building)

## 🔧 Configuration

Edit `configs/config.yaml` to customize:

```yaml
server:
  port: 8080
  enable_auth: true        # Enable JWT authentication

database:
  enabled: true           # Optional: Set to false to run without DB
  host: localhost
  port: 5432
  database: protei_monitoring
  user: protei
  password: secure_password
  ssl_mode: require

ingestion:
  sources:
    - type: pcap_file
      path: /usr/protei/Protei_Monitoring/ingest
      watch: true
      watch_interval: 5s

protocols:
  diameter:
    enabled: true
    applications: [S6a, Gx, Gy]
  map:
    enabled: true
    version: 3
  cap:
    enabled: true
  inap:
    enabled: true
  pfcp:
    enabled: true
  ngap:
    enabled: true

analytics:
  kpis:
    enabled: true
    procedures: [attach_4g, registration_5g, pdu_session_5g]

vendor_dictionaries:
  enabled: true
  vendors: [ericsson, huawei, zte, nokia]
  path: dictionaries/
```

### License Configuration

Generate your license file:

```bash
./bin/generate_license \
  --customer "Operator_Name" \
  --expiry "2030-12-31" \
  --mac "00:11:22:33:44:55" \
  --2g --3g --4g --5g \
  --map --cap --inap --diameter --gtp --pfcp --http \
  --max-subscribers 10000000 \
  --max-tps 10000 \
  --output configs/license.json
```

### Environment Variables

```bash
# Enable database (optional)
export DB_ENABLED=true

# JWT secret for authentication
export JWT_SECRET=your_secret_key_here

# License file path
export LICENSE_PATH=/usr/protei/Protei_Monitoring/configs/license.json
```

## 🚦 Usage Examples

### Process PCAP File

```bash
# Copy PCAP to ingestion directory
cp capture.pcap /usr/protei/Protei_Monitoring/ingest/

# Application automatically processes and generates:
# - Events: out/events/events_2025-11-13.jsonl
# - CDRs: out/cdr/cdr_2025-11-13_10.csv
# - Diagrams: out/diagrams/*.svg
```

### Query API

```bash
# Health check
curl http://localhost:8080/health

# Authenticate and get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

# Get license information
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/license | jq

# Get KPI report
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/kpi | jq

# Get active sessions
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/sessions | jq

# Get topology information
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/topology | jq

# Download PCAP for specific session
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/sessions/TID_123456/pcap" \
  -o session.pcap

# Prometheus metrics
curl http://localhost:8080/metrics

# Logout
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/auth/logout
```

## 📊 Output Formats

- **Events**: JSONL (one decoded message per line)
- **CDR**: CSV with configurable fields
- **Diagrams**: SVG (scalable vector graphics)
- **Logs**: JSON-formatted application logs
- **Metrics**: Prometheus-compatible format

## 🛠️ Development

```bash
# Clone repository
git clone https://github.com/protei/monitoring.git
cd monitoring

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run locally
make run
```

## 🔐 Security

### Authentication & Authorization
- **JWT-based authentication** with bcrypt password hashing
- **RBAC support** with predefined roles:
  - `admin`: Full system access
  - `engineer`: View sessions, KPIs, download PCAPs, create filters
  - `noc_viewer`: View-only access to dashboard and KPIs
  - `security_auditor`: Audit logs and alarm access
- **LDAP integration** (placeholder for enterprise directory)
- **Session management** with configurable token expiry

### License Management
- **MAC address binding**: License tied to specific network interface
- **HMAC-SHA256 signature** validation
- **Feature-level control**: Enable/disable protocols per license
- **Capacity limits**: Max subscribers and transactions per second (TPS)
- **Expiry enforcement**: Automatic validation on startup

### Network Security
- **Local-only mode** for sensitive environments
- **IP whitelisting** support
- **TLS/HTTPS** configuration options
- **Secure credential storage** with bcrypt

### Audit & Compliance
- **Audit log** for all administrative actions
- **Session tracking** with IP address logging
- **Failed login monitoring**
- **License validation logs**

## 🌟 Advantages

✅ **Self-contained**: Single binary, no external dependencies
✅ **Multi-vendor**: Support for all major equipment vendors
✅ **High performance**: Go-based concurrency, handles millions of messages
✅ **Production-ready**: Graceful shutdown, health checks, log rotation
✅ **Extensible**: YAML-based vendor dictionaries, plugin architecture
✅ **Complete solution**: Decode → Correlate → Analyze → Visualize
✅ **Enterprise security**: JWT authentication, RBAC, license management
✅ **Database integration**: Optional PostgreSQL with automatic migrations
✅ **PCAP processing**: File-based capture with directory monitoring
✅ **Comprehensive protocols**: All 2G/3G/4G/5G protocols in one platform

## 🎉 What's New in v2.0

### New Protocol Decoders
- **CAP (CAMEL Application Part)**: Full support for phases 1-4, prepaid/IN services
- **INAP (Intelligent Network Application Part)**: CS-1/CS-2/CS-3 with 36 operation codes
- **PFCP (Packet Forwarding Control Protocol)**: SMF-UPF communication on N4 interface
- **NGAP (NG Application Protocol)**: 5G gNB-AMF signaling with 51 procedure codes
- **S1AP**: 4G eNB-MME signaling with 48 procedure codes
- **NAS (Non-Access Stratum)**: 4G/5G UE-Core messaging with security header support

### Infrastructure Enhancements
- **PCAP Capture Engine**:
  - File-based ingestion with automatic directory monitoring
  - Multi-layer packet parsing (Ethernet → IP → TCP/UDP/SCTP → Application)
  - Metadata extraction for correlation

- **Database Layer**:
  - PostgreSQL integration with connection pooling
  - Liquibase-style migration system with 9 changesets
  - Tables: sessions, transactions, KPIs, topology, dictionaries, alarms, audit logs
  - Optional deployment (runs fine without database)

- **Authentication & Authorization**:
  - JWT token-based authentication with configurable expiry
  - Bcrypt password hashing for secure credential storage
  - RBAC with 4 predefined roles (admin, engineer, noc_viewer, security_auditor)
  - Session management with IP tracking
  - LDAP integration ready (placeholder implementation)

- **License Management**:
  - MAC address binding for hardware-locked licenses
  - HMAC-SHA256 cryptographic signature validation
  - Feature-level enablement (protocols, generations, capacity)
  - Subscriber and TPS limits enforcement
  - Expiry date validation on startup
  - License generation tool included

- **Vendor Dictionary System**:
  - YAML-based extensible format for AVPs/IEs
  - Support for Diameter, GTP, and cause code dictionaries
  - Per-vendor organization (Ericsson, Huawei, ZTE, Nokia)
  - Runtime loading and caching
  - Sample dictionaries included

### API Enhancements
- `/api/auth/login` - JWT token generation
- `/api/auth/logout` - Session invalidation
- `/api/license` - License information and status
- `/api/topology` - Network element topology
- `/api/sessions/:tid/pcap` - Download PCAP for specific session

### Operational Improvements
- License validation on startup with detailed error messages
- Protocol decoder registration based on license features
- Enhanced dashboard showing license status and feature availability
- Database health monitoring
- Authentication required for all API endpoints (except health check)

## 🗺️ Roadmap

### Completed in v2.0 ✅
- [x] PCAP file capture and processing
- [x] Database integration (PostgreSQL)
- [x] Authentication and authorization (JWT/RBAC)
- [x] License management system
- [x] Vendor dictionary support
- [x] CAP/INAP/PFCP/NGAP/S1AP/NAS protocol decoders

### Planned for Future Releases
- [ ] ML-based anomaly detection
- [ ] Live traffic capture (eBPF/SPAN/port mirroring)
- [ ] Kafka streaming integration
- [ ] Grafana dashboard templates
- [ ] 6G protocol readiness
- [ ] Distributed deployment support
- [ ] REST API rate limiting
- [ ] Multi-tenancy support
- [ ] Custom report builder
- [ ] Advanced LDAP/AD integration

---

**Protei_Monitoring v2.0** - Your complete telecom network intelligence platform 🚀
