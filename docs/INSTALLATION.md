# Protei_Monitoring Installation Guide

## Overview

Protei_Monitoring is a telecom-grade multi-protocol monitoring and analysis platform for 2G/3G/4G/5G networks.

## Requirements

### System Requirements

- **OS**: RHEL 9.6, Ubuntu 22.04, or compatible Linux distribution
- **CPU**: 4+ cores recommended (8+ for high-volume environments)
- **RAM**: 8GB minimum, 16GB+ recommended
- **Disk**: 100GB+ for logs and CDR storage
- **Go**: 1.21+ (for building from source)

### Network Requirements

- Access to network traffic capture (SPAN/TAP ports or PCAP files)
- HTTP port 8080 (default, configurable) for dashboard and API

## Installation Methods

### Method 1: Build and Install from Source

#### Step 1: Clone Repository

```bash
git clone https://github.com/protei/monitoring.git
cd monitoring
```

#### Step 2: Build

```bash
make all
```

This will:
- Download dependencies
- Build the binary
- Create build artifacts in `build/`

#### Step 3: Install

```bash
sudo make install
```

This installs to `/usr/protei/Protei_Monitoring/` with the following structure:

```
/usr/protei/Protei_Monitoring/
├── bin/
│   └── protei-monitoring          # Main binary
├── configs/
│   ├── config.yaml                # Main configuration
│   └── rules.yaml                 # Recommendation rules
├── logs/                          # Application logs
├── out/
│   ├── events/                    # Decoded message events (JSONL)
│   ├── cdr/                       # Call Detail Records (CSV)
│   └── diagrams/                  # Ladder diagrams (SVG)
├── ingest/                        # Input directory for PCAP files
├── scripts/
│   ├── start.sh                   # Start application
│   ├── stop.sh                    # Stop application
│   ├── restart.sh                 # Restart application
│   └── status.sh                  # Check status
└── dictionaries/
    ├── ericsson/                  # Ericsson vendor dictionaries
    ├── huawei/                    # Huawei vendor dictionaries
    ├── zte/                       # ZTE vendor dictionaries
    └── nokia/                     # Nokia vendor dictionaries
```

### Method 2: Binary Installation

If you have a pre-built binary:

```bash
# Create installation directory
sudo mkdir -p /usr/protei/Protei_Monitoring/{bin,configs,logs,out,ingest,scripts}

# Copy binary
sudo cp protei-monitoring /usr/protei/Protei_Monitoring/bin/

# Copy configuration files
sudo cp configs/*.yaml /usr/protei/Protei_Monitoring/configs/

# Copy control scripts
sudo cp scripts/*.sh /usr/protei/Protei_Monitoring/scripts/
sudo chmod +x /usr/protei/Protei_Monitoring/scripts/*.sh
```

## Configuration

### Main Configuration

Edit `/usr/protei/Protei_Monitoring/configs/config.yaml`:

```yaml
# Key settings to configure:

server:
  host: 0.0.0.0        # Listen on all interfaces
  port: 8080           # Dashboard and API port

ingestion:
  sources:
    - type: pcap_file
      path: /usr/protei/Protei_Monitoring/ingest
      watch: true      # Auto-process new files
      pattern: "*.pcap"

storage:
  logs:
    path: /usr/protei/Protei_Monitoring/logs
    level: info        # debug, info, warn, error
```

### Protocol Enablement

Enable/disable protocols as needed:

```yaml
protocols:
  map:
    enabled: true
  diameter:
    enabled: true
    applications:
      - S6a
      - Gx
      - Gy
  gtp:
    enabled: true
    versions: [1, 2]
  ngap:
    enabled: true
  s1ap:
    enabled: true
```

## Running the Application

### Start

```bash
sudo /usr/protei/Protei_Monitoring/scripts/start.sh
```

Output:
```
====================================
  Protei_Monitoring Start Script
====================================

Starting Protei_Monitoring...
✓ Protei_Monitoring started successfully (PID: 12345)
  Log file: /usr/protei/Protei_Monitoring/logs/console.log
  Dashboard: http://localhost:8080

Application is ready!
```

### Check Status

```bash
sudo /usr/protei/Protei_Monitoring/scripts/status.sh
```

### Stop

```bash
sudo /usr/protei/Protei_Monitoring/scripts/stop.sh
```

### Restart

```bash
sudo /usr/protei/Protei_Monitoring/scripts/restart.sh
```

## Accessing the Dashboard

Once started, access the web dashboard at:

```
http://<server-ip>:8080
```

### API Endpoints

- **Health Check**: `http://localhost:8080/health`
- **Metrics**: `http://localhost:8080/metrics` (Prometheus-compatible)
- **Sessions**: `http://localhost:8080/api/sessions`
- **KPI Report**: `http://localhost:8080/api/kpi`
- **Roaming Data**: `http://localhost:8080/api/roaming`

## Usage Examples

### Processing PCAP Files

1. Copy PCAP files to the ingestion directory:

```bash
sudo cp capture.pcap /usr/protei/Protei_Monitoring/ingest/
```

2. The application will automatically:
   - Detect and process the file
   - Decode all supported protocols
   - Correlate messages into sessions
   - Calculate KPIs
   - Generate CDRs and diagrams

### Viewing Results

**Events (decoded messages)**:
```bash
cat /usr/protei/Protei_Monitoring/out/events/events_2025-11-13.jsonl
```

**CDR (Call Detail Records)**:
```bash
cat /usr/protei/Protei_Monitoring/out/cdr/cdr_2025-11-13_10.csv
```

**Ladder Diagrams**:
```bash
ls /usr/protei/Protei_Monitoring/out/diagrams/
```

## Monitoring and Logs

### Application Logs

```bash
tail -f /usr/protei/Protei_Monitoring/logs/app.log
```

### Console Output

```bash
tail -f /usr/protei/Protei_Monitoring/logs/console.log
```

### Health Monitoring

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "healthy": true,
  "uptime": 3600,
  "messages": 15234
}
```

## Troubleshooting

### Application Won't Start

1. Check if port 8080 is available:
   ```bash
   sudo lsof -i :8080
   ```

2. Check logs:
   ```bash
   cat /usr/protei/Protei_Monitoring/logs/console.log
   ```

3. Verify permissions:
   ```bash
   ls -la /usr/protei/Protei_Monitoring/bin/protei-monitoring
   ```

### No PCAP Files Being Processed

1. Verify ingestion path in config:
   ```bash
   grep -A5 "ingestion:" /usr/protei/Protei_Monitoring/configs/config.yaml
   ```

2. Check file permissions:
   ```bash
   ls -la /usr/protei/Protei_Monitoring/ingest/
   ```

3. Check logs for decode errors:
   ```bash
   grep ERROR /usr/protei/Protei_Monitoring/logs/app.log
   ```

### High Memory Usage

1. Adjust cache sizes in config:
   ```yaml
   correlation:
     tid_cache_size: 500000  # Reduce if needed
   ```

2. Enable memory limits:
   ```yaml
   performance:
     max_memory_mb: 4096
   ```

## Systemd Service (Optional)

Create `/etc/systemd/system/protei-monitoring.service`:

```ini
[Unit]
Description=Protei Monitoring Service
After=network.target

[Service]
Type=forking
User=root
Group=root
WorkingDirectory=/usr/protei/Protei_Monitoring
ExecStart=/usr/protei/Protei_Monitoring/scripts/start.sh
ExecStop=/usr/protei/Protei_Monitoring/scripts/stop.sh
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable protei-monitoring
sudo systemctl start protei-monitoring
sudo systemctl status protei-monitoring
```

## Performance Tuning

For high-volume environments:

```yaml
ingestion:
  workers: 16           # Increase for more cores
  buffer_size: 200000   # Larger buffer

performance:
  max_goroutines: 20000
  numa_aware: true
  async_processing: true

correlation:
  tid_cache_size: 2000000
```

## Security Considerations

1. **Firewall**: Restrict access to port 8080
2. **Authentication**: Enable in production:
   ```yaml
   security:
     auth_enabled: true
     auth_type: jwt
   ```

3. **Local-only mode**:
   ```yaml
   security:
     local_only: true
   ```

## Support

For issues and questions:
- GitHub Issues: https://github.com/protei/monitoring/issues
- Documentation: https://docs.protei.com/monitoring

## Next Steps

- Configure vendor dictionaries for your equipment
- Set up automated PCAP capture
- Integrate with monitoring systems (Grafana, ELK)
- Configure alert thresholds in `rules.yaml`
