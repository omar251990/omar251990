# Protei_Monitoring

🌐 **Full Telecom-Grade Multi-Protocol Monitoring & Analysis Platform**

A comprehensive monitoring solution for 2G/3G/4G/5G networks with deep protocol decoding, intelligent correlation, KPI analytics, and real-time visualization.

## 📋 Overview

Protei_Monitoring is a carrier-grade platform capable of:

- **Multi-Protocol Decoding**: MAP, CAP, INAP, Diameter, GTP-C, PFCP, HTTP, NGAP, S1AP, NAS
- **Intelligent Correlation**: Automatic session tracking with unique Transaction IDs (TID)
- **KPI Analytics**: Success rates, latency metrics, failure analysis, cause distribution
- **Roaming Intelligence**: Inbound/outbound roamer tracking with cell-level heatmaps
- **Real-time Visualization**: Ladder diagrams, network flow graphs, interactive dashboard
- **Vendor Support**: Ericsson, Huawei, ZTE, Nokia equipment with extensible dictionaries
- **Production-Ready**: Self-contained binary, graceful shutdown, health monitoring, automatic log rotation

## 🚀 Quick Start

### Build and Install

```bash
# Build the application
make all

# Install to /usr/protei/Protei_Monitoring
sudo make install

# Start the service
sudo /usr/protei/Protei_Monitoring/scripts/start.sh
```

### Access Dashboard

Open your browser to: `http://localhost:8080`

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

ingestion:
  sources:
    - type: pcap_file
      path: /usr/protei/Protei_Monitoring/ingest
      watch: true

protocols:
  diameter:
    enabled: true
    applications: [S6a, Gx, Gy]

analytics:
  kpis:
    enabled: true
    procedures: [attach_4g, registration_5g]
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

# Get KPI report
curl http://localhost:8080/api/kpi | jq

# Get active sessions
curl http://localhost:8080/api/sessions

# Prometheus metrics
curl http://localhost:8080/metrics
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

- Optional authentication (Basic, OAuth2, JWT)
- RBAC support
- Local-only mode for sensitive environments
- Configurable IP whitelisting

## 🌟 Advantages

✅ **Self-contained**: Single binary, no external dependencies
✅ **Multi-vendor**: Support for all major equipment vendors
✅ **High performance**: Go-based concurrency, handles millions of messages
✅ **Production-ready**: Graceful shutdown, health checks, log rotation
✅ **Extensible**: YAML-based vendor dictionaries, plugin architecture
✅ **Complete solution**: Decode → Correlate → Analyze → Visualize

## 🗺️ Roadmap

- [ ] ML-based anomaly detection
- [ ] Live traffic capture (eBPF/SPAN)
- [ ] Kafka streaming integration
- [ ] Grafana dashboard templates
- [ ] 6G protocol readiness
- [ ] Distributed deployment support

---

**Protei_Monitoring** - Your complete telecom network intelligence platform 🚀
