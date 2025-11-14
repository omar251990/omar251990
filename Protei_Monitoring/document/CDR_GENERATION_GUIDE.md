# Protei Monitoring CDR Generation System

## Overview

The CDR (Call Detail Record) generation system provides comprehensive recording of all signaling and user plane transactions across 10 supported protocols. CDRs are automatically generated, rotated, compressed, and tracked in the database for audit and billing purposes.

## Table of Contents

1. [Architecture](#architecture)
2. [Supported Protocols](#supported-protocols)
3. [CDR Formats](#cdr-formats)
4. [File Organization](#file-organization)
5. [Rotation and Compression](#rotation-and-compression)
6. [Database Tracking](#database-tracking)
7. [Management Tools](#management-tools)
8. [Integration Examples](#integration-examples)
9. [Performance Considerations](#performance-considerations)

---

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                      CDR Manager                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │   MAP    │  │   CAP    │  │  Diameter│  │   GTP    │  │
│  │  Writer  │  │  Writer  │  │  Writer  │  │  Writer  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  PFCP    │  │  HTTP/2  │  │   NGAP   │  │  S1AP    │  │
│  │  Writer  │  │  Writer  │  │  Writer  │  │  Writer  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
│  ┌──────────┐  ┌──────────┐                               │
│  │   NAS    │  │   INAP   │                               │
│  │  Writer  │  │  Writer  │                               │
│  └──────────┘  └──────────┘                               │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              File System + Database Tracker                 │
│                                                             │
│  /usr/protei/Protei_Monitoring/cdr/                        │
│  ├── MAP/                                                   │
│  ├── CAP/                                                   │
│  ├── INAP/                                                  │
│  ├── Diameter/                                              │
│  ├── GTP/                                                   │
│  ├── PFCP/                                                  │
│  ├── HTTP2/                                                 │
│  ├── NGAP/                                                  │
│  ├── S1AP/                                                  │
│  ├── NAS/                                                   │
│  └── combined/                                              │
└─────────────────────────────────────────────────────────────┘
```

### Key Features

- **Protocol-Specific Writers**: Separate CDR writer for each protocol
- **Automatic Rotation**: Size-based and time-based rotation
- **Compression**: gzip compression of rotated files
- **Database Tracking**: Metadata stored in PostgreSQL
- **Thread-Safe**: Concurrent writes from multiple decoders
- **Multiple Formats**: CSV, JSON, XML support

---

## Supported Protocols

### 1. MAP (2G/3G Signaling)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Operation Type, Operation Code, Invoke ID
- Result, Result Code, Duration (ms)
- SCCP Called/Calling Party
- MCC, MNC, LAC, Cell ID

### 2. CAP (CAMEL Application Part)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Service Key, Calling Party, Called Party
- Operation Type, Event Type
- Result, Result Code, Duration (ms)
- SCCP Called/Calling Party

### 3. INAP (Intelligent Network Application Part)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Service Key, Calling Party, Called Party
- Operation Type, Trigger Type
- Result, Result Code, Duration (ms)

### 4. Diameter (4G/5G Core)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Session ID, Command Code, Application ID
- Operation Type, Result, Result Code, Duration (ms)
- Origin Host/Realm, Destination Host/Realm

### 5. GTP (User Plane - 3G/4G)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Message Type, TEID, Sequence Number
- APN, PDN Type
- Source IP, Destination IP
- Bytes Uplink/Downlink

### 6. PFCP (5G User Plane Control)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Message Type, SEID, Node ID, FSE ID
- Result, Result Code, Duration (ms)
- Uplink/Downlink Bytes

### 7. HTTP/2 (5G Service-Based Architecture)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Method, URI, Status Code
- Service Name, API Version
- Source NF, Target NF
- Duration (ms)

### 8. NGAP (5G RAN)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Procedure Code, AMF UE ID, RAN UE ID
- Global RAN ID, GUAMI, Cause
- Result, Result Code, Duration (ms)

### 9. S1AP (4G RAN)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Procedure Code, MME UE ID, eNB UE ID
- TAI, E-UTRAN CGI, Cause
- Result, Result Code, Duration (ms)

### 10. NAS (4G/5G Mobile)
**Fields:**
- Timestamp, Transaction ID, IMSI, MSISDN
- Message Type, Security Header
- Protocol Discriminator, EPS Mobile Identity
- EMM Cause, ESM Cause
- Duration (ms)

---

## CDR Formats

### CSV Format (Default)

**Advantages:**
- Easy to import into Excel/databases
- Human-readable
- Small file size

**Example - MAP CDR:**
```csv
Timestamp,Protocol,TransactionID,IMSI,MSISDN,OperationType,Result,ResultCode,DurationMs,OperationCode,InvokeID,SCCP_Called,SCCP_Calling,MCC,MNC,LAC,CellID
2024-01-15T10:30:45Z,MAP,TXN001,234150123456789,1234567890,UpdateLocation,Success,0,150,2,1,12345678,87654321,234,15,1001,A1B2
2024-01-15T10:31:12Z,MAP,TXN002,234150987654321,9876543210,SendAuthInfo,Success,0,85,56,2,12345678,87654321,234,15,1001,A1B2
```

### JSON Format

**Advantages:**
- Structured data
- Easy to parse programmatically
- Supports nested objects

**Example - Diameter CDR:**
```json
{"timestamp":"2024-01-15T10:30:45Z","protocol":"Diameter","transaction_id":"TXN001","imsi":"234150123456789","msisdn":"1234567890","session_id":"sess-12345","command_code":316,"application_id":16777251,"operation_type":"ULR","result":"Success","result_code":2001,"duration_ms":120,"origin_host":"mme01.operator.com","origin_realm":"operator.com","destination_host":"hss01.operator.com","destination_realm":"operator.com"}
{"timestamp":"2024-01-15T10:31:00Z","protocol":"Diameter","transaction_id":"TXN002","imsi":"234150987654321","msisdn":"9876543210","session_id":"sess-12346","command_code":318,"application_id":16777251,"operation_type":"AIR","result":"Success","result_code":2001,"duration_ms":95,"origin_host":"mme01.operator.com","origin_realm":"operator.com","destination_host":"hss01.operator.com","destination_realm":"operator.com"}
```

---

## File Organization

### Directory Structure

```
/usr/protei/Protei_Monitoring/cdr/
├── MAP/
│   ├── MAP_20240115_103045.csv
│   ├── MAP_20240115_120000.csv.gz
│   └── MAP_20240115_130000.csv.gz
├── CAP/
│   ├── CAP_20240115_103045.csv
│   └── CAP_20240115_120000.csv.gz
├── Diameter/
│   ├── Diameter_20240115_103045.csv
│   └── Diameter_20240115_120000.csv.gz
├── GTP/
│   ├── GTP_20240115_103045.csv
│   └── GTP_20240115_120000.csv.gz
└── combined/
    └── All_Protocols_20240115.csv
```

### Naming Convention

**Format:** `{Protocol}_{YYYYMMDD}_{HHMMSS}.{extension}[.gz]`

**Examples:**
- `MAP_20240115_103045.csv` - Active MAP CDR file
- `Diameter_20240115_120000.csv.gz` - Compressed Diameter CDR
- `GTP_20240115_000000.json` - GTP CDR in JSON format

---

## Rotation and Compression

### Rotation Triggers

1. **Size-Based Rotation**
   - Default: 100 MB per file
   - Configurable in `config/system.cfg`

2. **Time-Based Rotation**
   - Default: 1 hour
   - Daily rotation: 24 hours
   - Configurable interval

### Rotation Process

1. Current file is closed and flushed
2. Metadata tracked in database (`cdr_files` table)
3. New file created with timestamp
4. CSV header written (if CSV format)
5. Old file compressed in background (if enabled)

### Compression

**Algorithm:** gzip -9 (maximum compression)

**Compression Ratios:**
- CSV files: 85-95% reduction
- JSON files: 80-90% reduction

**Example:**
```bash
# Original file
MAP_20240115_120000.csv       102 MB

# After compression
MAP_20240115_120000.csv.gz     12 MB  (88% reduction)
```

### Configuration

**File:** `/usr/protei/Protei_Monitoring/config/system.cfg`

```bash
# CDR Configuration
CDR_FORMAT="csv"               # csv, json, xml
CDR_ROTATION_SIZE_MB=100       # Rotate at 100 MB
CDR_ROTATION_HOURS=1           # Rotate every 1 hour
CDR_COMPRESSION=true           # Enable compression
CDR_RETENTION_DAYS=90          # Keep CDRs for 90 days
```

---

## Database Tracking

### CDR Files Table

All CDR files are tracked in the `cdr_files` table:

```sql
CREATE TABLE cdr_files (
    id BIGSERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    record_count BIGINT DEFAULT 0,
    file_size_bytes BIGINT DEFAULT 0,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    compressed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Queries

**List all CDR files for a protocol:**
```sql
SELECT
    filename,
    record_count,
    pg_size_pretty(file_size_bytes) as size,
    start_time,
    end_time
FROM cdr_files
WHERE protocol = 'MAP'
ORDER BY start_time DESC
LIMIT 100;
```

**Get CDR statistics by protocol:**
```sql
SELECT
    protocol,
    COUNT(*) as file_count,
    SUM(record_count) as total_records,
    pg_size_pretty(SUM(file_size_bytes)) as total_size
FROM cdr_files
WHERE start_time >= NOW() - INTERVAL '7 days'
GROUP BY protocol
ORDER BY protocol;
```

**Find CDRs for a specific time range:**
```sql
SELECT filename, record_count
FROM cdr_files
WHERE protocol = 'Diameter'
  AND start_time >= '2024-01-15 10:00:00'
  AND end_time <= '2024-01-15 12:00:00'
ORDER BY start_time;
```

---

## Management Tools

### CDR Management Script

**Location:** `/usr/protei/Protei_Monitoring/scripts/utils/manage_cdr.sh`

### Commands

#### 1. List CDR Files

```bash
# List all CDR files
sudo scripts/utils/manage_cdr.sh list

# List MAP CDR files only
sudo scripts/utils/manage_cdr.sh list MAP

# List files for specific date
sudo scripts/utils/manage_cdr.sh list MAP 20240115
```

**Output:**
```
Protocol: MAP
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename                              Size    Date
────────────────────────────────────────────────────────────
MAP_20240115_103045.csv              98M     2024-01-15
MAP_20240115_120000.csv.gz           11M     2024-01-15
MAP_20240115_130000.csv.gz           12M     2024-01-15

Total files: 3
Compressed: 2
Total size: 121 MB
```

#### 2. Compress Old Files

```bash
# Compress all files older than 1 day
sudo scripts/utils/manage_cdr.sh compress 1

# Compress only GTP files older than 7 days
sudo scripts/utils/manage_cdr.sh compress 7 GTP
```

**Output:**
```
✅ Compressed: MAP_20240114_103045.csv (saved 89 MB)
✅ Compressed: MAP_20240114_120000.csv (saved 91 MB)

Compression complete!
Files compressed: 2
Space saved: 180 MB
```

#### 3. Cleanup Old Files

```bash
# Delete files older than 90 days (dry run)
sudo scripts/utils/manage_cdr.sh cleanup 90 "" yes

# Actually delete files older than 90 days
sudo scripts/utils/manage_cdr.sh cleanup 90

# Delete only Diameter files older than 30 days
sudo scripts/utils/manage_cdr.sh cleanup 30 Diameter
```

#### 4. Generate Statistics

```bash
# Statistics for all protocols
sudo scripts/utils/manage_cdr.sh stats

# Statistics for Diameter only
sudo scripts/utils/manage_cdr.sh stats Diameter
```

**Output:**
```
Protocol  Files  Total Size  Compressed  Oldest        Newest
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MAP       245    28 GB       240         2023-10-15    2024-01-15
CAP       180    15 GB       175         2023-10-15    2024-01-15
Diameter  420    52 GB       415         2023-10-15    2024-01-15
GTP       890    125 GB      885         2023-10-15    2024-01-15
HTTP2     350    42 GB       345         2023-10-15    2024-01-15
NGAP      280    35 GB       275         2023-10-15    2024-01-15

Overall Statistics:
  Total CDR files: 2,365
  Total size: 297 GB
  Compressed files: 2,335
  Compression ratio: 98%
```

#### 5. Export CDR Data

```bash
# Export Diameter CDRs to CSV
sudo scripts/utils/manage_cdr.sh export Diameter /tmp/diameter_export.csv
```

#### 6. Verify File Integrity

```bash
# Verify all compressed files
sudo scripts/utils/manage_cdr.sh verify

# Verify only MAP files
sudo scripts/utils/manage_cdr.sh verify MAP
```

---

## Integration Examples

### Example 1: Writing MAP CDRs

```go
package main

import (
    "database/sql"
    "time"
    "protei/pkg/cdr"
)

func main() {
    // Initialize database
    db, _ := sql.Open("postgres", "...")

    // Create CDR manager
    rotation := cdr.CDRRotation{
        MaxSizeMB:   100,
        MaxDuration: 1 * time.Hour,
        Compress:    true,
    }

    manager, _ := cdr.NewCDRManager(
        "/usr/protei/Protei_Monitoring/cdr",
        cdr.FormatCSV,
        rotation,
        db,
    )
    defer manager.Close()

    // Write MAP CDR
    mapCDR := &cdr.MapCDR{
        Timestamp:     time.Now(),
        TransactionID: "TXN001",
        IMSI:          "234150123456789",
        MSISDN:        "1234567890",
        OperationType: "UpdateLocation",
        OperationCode: 2,
        Result:        "Success",
        ResultCode:    0,
        DurationMs:    150,
        SCCP_Called:   "12345678",
        SCCP_Calling:  "87654321",
        MCC:           "234",
        MNC:           "15",
        LAC:           "1001",
        CellID:        "A1B2",
    }

    manager.WriteMapCDR(mapCDR)
}
```

### Example 2: Writing Diameter CDRs

```go
// Write Diameter CDR
diameterCDR := &cdr.DiameterCDR{
    Timestamp:        time.Now(),
    TransactionID:    "TXN002",
    IMSI:             "234150123456789",
    MSISDN:           "1234567890",
    SessionID:        "sess-12345",
    CommandCode:      316, // ULR
    ApplicationID:    16777251, // S6a
    OperationType:    "UpdateLocation",
    Result:           "Success",
    ResultCode:       2001,
    DurationMs:       120,
    OriginHost:       "mme01.operator.com",
    OriginRealm:      "operator.com",
    DestinationHost:  "hss01.operator.com",
    DestinationRealm: "operator.com",
}

manager.WriteDiameterCDR(diameterCDR)
```

### Example 3: Writing GTP CDRs

```go
// Write GTP CDR
gtpCDR := &cdr.GtpCDR{
    Timestamp:     time.Now(),
    TransactionID: "TXN003",
    IMSI:          "234150123456789",
    MSISDN:        "1234567890",
    MessageType:   "CreateSessionResponse",
    TEID:          0x12345678,
    SequenceNumber: 100,
    APN:           "internet",
    PDN_Type:      "IPv4",
    Result:        "Success",
    ResultCode:    0,
    DurationMs:    200,
    SourceIP:      "10.1.1.1",
    DestIP:        "10.2.2.2",
    BytesUplink:   1024000,
    BytesDownlink: 5120000,
}

manager.WriteGtpCDR(gtpCDR)
```

---

## Performance Considerations

### Write Performance

- **Throughput**: 50,000+ CDRs/second per protocol
- **Latency**: < 1 ms per write (non-blocking)
- **Buffering**: 4096-byte buffer per writer

### Optimization Tips

1. **Use CSV format** for maximum performance
2. **Enable compression** to save disk space (80-90% savings)
3. **Adjust rotation size** based on traffic:
   - Low traffic: 50 MB
   - Medium traffic: 100 MB
   - High traffic: 500 MB
4. **Separate disk** for CDR storage if high TPS
5. **Regular cleanup** of old CDRs (90 days default)

### Disk Space Estimation

**Formula:**
```
Daily Disk Space = (TPS × Avg Record Size × 86400) × (1 - Compression Ratio)
```

**Example Calculation:**
- TPS: 1,000
- Avg Record Size: 500 bytes
- Compression Ratio: 90%

```
Daily Space = (1000 × 500 × 86400) × 0.10
            = 43,200,000,000 × 0.10
            = 4.32 GB per day (compressed)
```

**90-Day Retention:**
```
Total Storage = 4.32 GB × 90 = 388 GB
```

### Monitoring

**Check CDR writer statistics:**
```bash
# Via API
curl http://localhost:8080/api/v1/cdr/stats

# Response:
{
  "MAP": {
    "record_count": 125000,
    "bytes_written": 62500000,
    "duration": "3h45m"
  },
  "Diameter": {
    "record_count": 250000,
    "bytes_written": 125000000,
    "duration": "3h45m"
  }
}
```

---

## Troubleshooting

### Issue 1: CDR files not being created

**Symptoms:** No CDR files in `/usr/protei/Protei_Monitoring/cdr/`

**Diagnosis:**
```bash
# Check directory permissions
ls -la /usr/protei/Protei_Monitoring/cdr/

# Check application logs
tail -f /usr/protei/Protei_Monitoring/logs/application/protei-monitoring.log | grep CDR
```

**Solution:**
```bash
# Fix permissions
sudo chown -R protei:protei /usr/protei/Protei_Monitoring/cdr/
sudo chmod -R 755 /usr/protei/Protei_Monitoring/cdr/
```

### Issue 2: Disk space running out

**Symptoms:** "No space left on device" errors

**Diagnosis:**
```bash
# Check disk usage
df -h /usr/protei/

# Check CDR directory size
du -sh /usr/protei/Protei_Monitoring/cdr/*
```

**Solution:**
```bash
# Compress old files
sudo scripts/utils/manage_cdr.sh compress 1

# Delete old files
sudo scripts/utils/manage_cdr.sh cleanup 30
```

### Issue 3: Corrupted compressed files

**Symptoms:** Cannot decompress .gz files

**Diagnosis:**
```bash
# Verify all compressed files
sudo scripts/utils/manage_cdr.sh verify
```

**Solution:**
```bash
# If corruption detected, check original file
# Compressed files are created in background
# Original file is deleted only after successful compression
```

---

## Summary

The Protei Monitoring CDR system provides:

✅ **Complete Coverage** - 10 protocols supported
✅ **Automatic Rotation** - Size and time-based
✅ **Compression** - 80-90% space savings
✅ **Database Tracking** - Full audit trail
✅ **Management Tools** - Easy administration
✅ **High Performance** - 50K+ CDRs/sec per protocol
✅ **Multiple Formats** - CSV, JSON, XML
✅ **Thread-Safe** - Concurrent writes supported

For support or questions, refer to the main documentation or contact the development team.
