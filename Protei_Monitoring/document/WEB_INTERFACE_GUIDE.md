# Protei Monitoring v2.0 - Web Interface Guide

Complete guide to using the Protei Monitoring web interface.

## Table of Contents

1. [Dashboard Overview](#dashboard-overview)
2. [Navigation](#navigation)
3. [Features](#features)
4. [Search and Filtering](#search-and-filtering)
5. [Visualization](#visualization)
6. [AI Features](#ai-features)
7. [Administration](#administration)

---

## Dashboard Overview

### Main Dashboard

The dashboard provides a real-time overview of your telecom monitoring system.

**Key Sections:**

1. **System Status** (Top Bar)
   - Application status (Running/Stopped)
   - Uptime counter
   - CPU and memory usage
   - Active capture interfaces

2. **Protocol Status** (Left Panel)
   ```
   ✅ MAP      - Active (1,234 messages)
   ✅ CAP      - Active (567 messages)
   ✅ INAP     - Active (890 messages)
   ✅ Diameter - Active (45,678 messages)
   ✅ GTP      - Active (123,456 messages)
   ✅ PFCP     - Active (34,567 messages)
   ✅ HTTP/2   - Active (8,901 messages)
   ✅ NGAP     - Active (12,345 messages)
   ✅ S1AP     - Active (9,876 messages)
   ✅ NAS      - Active (11,234 messages)
   ```

3. **Real-Time Statistics** (Center Panel)
   - Messages per second (live graph)
   - Active sessions count
   - Success rate percentage
   - Error rate percentage

4. **Recent Sessions** (Bottom Panel)
   - Last 20 sessions with timestamps
   - Protocol type
   - Source/Destination
   - Status (Success/Failed/Ongoing)
   - Quick actions (View/Export/Analyze)

---

## Navigation

### Main Menu

**Left Sidebar Navigation:**

#### 🏠 Dashboard
- System overview
- Real-time statistics
- Quick actions

#### 📊 Monitoring
- **Sessions** - Active and historical sessions
  - View all sessions
  - Filter by protocol
  - Search by identifiers
  - Session details with ladder diagram

- **Protocols** - Protocol-specific views
  - MAP sessions (Location Update, SMS, etc.)
  - CAP sessions (CAMEL interactions)
  - Diameter sessions (Authentication, Charging)
  - GTP tunnels (Data sessions)
  - And more...

- **Subscribers** - Subscriber-centric view
  - Search by IMSI/MSISDN/IMEI
  - Subscriber timeline
  - Location history
  - Active sessions

#### 🤖 AI & Intelligence
- **Analysis** - AI-based issue detection
  - Recent issues
  - Issue statistics
  - Root cause analysis
  - Recommendations

- **Knowledge Base** - Protocol references
  - 3GPP Standards (18 documents)
  - Error codes reference
  - Procedure descriptions
  - Search functionality

- **Flow Reconstruction** - Message flow analysis
  - Standard procedures (5 templates)
  - Deviation detection
  - Dual view (Actual vs. Expected)
  - Completeness score

#### 📈 Analytics
- **KPIs** - Key Performance Indicators
  - Success rate trends
  - Average response time
  - Error distribution
  - Protocol usage

- **Reports** - Pre-built and custom reports
  - Daily summary reports
  - Protocol statistics
  - Subscriber activity
  - Custom report builder

#### ⚙️ Settings
- **Configuration** - System configuration
  - Protocol settings
  - Capture configuration
  - Database settings
  - Performance tuning

- **Users** - User management
  - Create/edit/delete users
  - Role assignment (Admin/Operator/Viewer)
  - Password policies
  - LDAP configuration

- **Security** - Security settings
  - Authentication settings
  - Session timeout
  - Audit logs
  - License information

- **System** - System administration
  - Service control (Start/Stop/Restart)
  - Log management
  - Backup/Restore
  - Update management

---

## Features

### 1. Session Monitoring

#### View All Sessions

**Location:** Monitoring → Sessions

**Features:**
- Paginated table with 50 sessions per page
- Real-time updates every 5 seconds
- Column sorting (click headers)
- Multi-column filtering

**Columns:**
- Transaction ID
- Timestamp (start/end)
- Protocol
- IMSI / MSISDN / IMEI
- Source → Destination
- Status
- Duration
- Message Count
- Actions (View/Export/Delete)

#### Session Details

Click any session to view:

1. **Session Summary**
   ```
   Transaction ID: TID-123456789
   Protocol: MAP
   Started: 2025-11-14 10:23:45
   Duration: 2.3 seconds
   Messages: 8
   Status: ✅ Success
   ```

2. **Message List**
   - All messages in chronological order
   - Direction indicators (→ ←)
   - Message types
   - Timestamp offsets
   - Decoded parameters

3. **Ladder Diagram**
   - Visual message flow
   - Network elements (MSC, HLR, VLR, etc.)
   - Message sequence
   - Timing information
   - Click messages for details

4. **Decoded Data**
   - Full protocol decode
   - Hex dump view
   - Parameter tree
   - Copy/export options

---

### 2. Subscriber Tracking

#### Search Subscribers

**Location:** Monitoring → Subscribers

**Search Options:**
- **IMSI**: 123456789012345
- **MSISDN**: +1234567890
- **IMEI**: 123456789012345
- **Session ID**: TID-xxx

#### Subscriber Profile

**Overview:**
```
IMSI: 123456789012345
MSISDN: +1234567890
IMEI: 123456789012345

Current Location:
  MCC: 123
  MNC: 45
  LAC: 1234
  Cell ID: 56789

Status: Active
Last Seen: 2025-11-14 10:30:15
First Seen: 2025-11-01 08:15:22
```

**Timeline View:**

Visual timeline showing:
- 📍 Location updates
- 📞 Voice calls
- 📱 Data sessions
- 💬 SMS messages
- ⚠️ Errors/Issues

Click any event to see full details.

**Location History:**

Map view (if configured) or table:
```
Timestamp           Location              Event
2025-11-14 10:30    Cell 56789 (LAC 1234) Location Update
2025-11-14 09:15    Cell 56788 (LAC 1234) Handover
2025-11-14 08:00    Cell 56780 (LAC 1230) Attach
```

**Active Sessions:**

All currently active sessions for this subscriber.

---

### 3. AI Analysis

#### Issue Detection

**Location:** AI & Intelligence → Analysis

**Dashboard shows:**

1. **Recent Issues** (Last 24 hours)
   ```
   ⚠️  High: 5 issues
   🔶  Medium: 23 issues
   ℹ️  Low: 67 issues
   ```

2. **Issue List**

Each issue shows:
- **Severity**: High/Medium/Low
- **Category**: Protocol Error, Network Issue, Performance, etc.
- **Description**: Human-readable issue summary
- **Affected Sessions**: Count and links
- **First/Last Occurrence**
- **Root Cause**: AI-generated analysis
- **Recommendations**: Suggested fixes

**Example Issue:**
```
⚠️ High Severity - Location Update Failures

Category: Protocol Error
Affected: 145 sessions (IMSI prefix: 12345678*)
Pattern: MAP Update Location Request → Reject (Cause: Unknown Subscriber)

Root Cause Analysis:
The HLR is consistently rejecting location updates for subscribers
with IMSI prefix 12345678*. This suggests a provisioning issue
in the HLR database for this subscriber range.

Recommendations:
1. Verify HLR provisioning for IMSI range 12345678*
2. Check HLR database consistency
3. Review recent HLR updates/migrations
4. Contact vendor if issue persists

Related Standards: 3GPP TS 29.002 (MAP)
```

#### Statistics Dashboard

Charts showing:
- Issue trend (last 7 days)
- Issue breakdown by category
- Issue breakdown by protocol
- Top affected subscribers
- Resolution status

---

### 4. Knowledge Base

#### Standards Library

**Location:** AI & Intelligence → Knowledge Base → Standards

**18 Built-in Standards:**

**3GPP Standards (12):**
1. TS 29.002 - MAP Protocol
2. TS 29.078 - CAP Protocol
3. TS 29.274 - GTPv2-C
4. TS 29.244 - PFCP
5. TS 29.272 - Diameter SWm/SWx
6. TS 29.273 - Diameter S6a/S6d
7. TS 38.413 - NGAP
8. TS 36.413 - S1AP
9. TS 24.301 - NAS (EPS)
10. TS 24.501 - NAS (5GS)
11. TS 29.500 - 5G SBI
12. TS 23.401 - EPC Architecture

**IETF RFCs (6):**
1. RFC 3261 - SIP
2. RFC 6733 - Diameter Base
3. RFC 7540 - HTTP/2
4. RFC 4960 - SCTP
5. RFC 3588 - Diameter (obsoleted by 6733)
6. RFC 2234 - ABNF Syntax

**Features:**
- Full-text search
- Quick reference lookup
- Hyperlinked sections
- Download PDF (if available)

#### Error Codes Reference

Search or browse error codes:

```
Protocol: MAP
Error Code: 0x01 (Unknown Subscriber)

Description:
This error indicates that the subscriber identity (IMSI) is not
recognized by the HLR. The subscriber may not be provisioned.

Typical Causes:
- Subscriber not provisioned in HLR
- IMSI typo in request
- Database synchronization issue

Troubleshooting:
1. Verify subscriber exists in HLR
2. Check IMSI format (15 digits)
3. Review HLR logs
4. Check database replication status

Standard Reference: 3GPP TS 29.002 Section 7.6.1
```

#### Procedure Reference

**5 Standard Procedures:**

1. **4G Attach Procedure**
   - Message sequence
   - Expected parameters
   - Success criteria
   - Common failures

2. **5G Registration**
   - Initial registration
   - Periodic registration
   - Mobility registration

3. **PDU Session Establishment**
   - Session creation flow
   - QoS negotiation
   - Failure scenarios

4. **GTP Tunnel Creation**
   - Create Session Request/Response
   - Modify Bearer Request/Response
   - Delete Session Request/Response

5. **MAP Location Update**
   - Update Location Request
   - Insert Subscriber Data
   - Update Location Response

Each procedure shows:
- Detailed message flow
- Required parameters
- Optional parameters
- Success/failure conditions
- Timing expectations

---

### 5. Flow Reconstruction

#### Analyze Message Flow

**Location:** AI & Intelligence → Flow Reconstruction

**Steps:**

1. **Select Procedure Template**
   - Choose from 5 standard procedures
   - Or use "Auto-detect" mode

2. **Upload/Select Messages**
   - Upload PCAP file
   - Or select from captured sessions
   - Filter by time range

3. **Run Analysis**

**Results:**

1. **Dual View Comparison**

   Left: **Actual Flow**
   ```
   UE → AMF: Registration Request
   AMF → AUSF: Auth Request
   AUSF → AMF: Auth Response
   AMF → UE: Registration Accept
   ```

   Right: **Expected Flow (Standard)**
   ```
   UE → AMF: Registration Request
   AMF → AUSF: Auth Request
   AUSF → UDM: Get Auth Data    ❌ MISSING
   UDM → AUSF: Auth Data        ❌ MISSING
   AUSF → AMF: Auth Response
   AMF → UE: Security Mode Cmd  ❌ MISSING
   UE → AMF: Security Mode Cpl  ❌ MISSING
   AMF → UE: Registration Accept
   ```

2. **Deviation Analysis**
   ```
   ⚠️ 3 deviations detected:

   1. Missing: AUSF → UDM Authentication Data Request
      Impact: Medium
      Note: May indicate direct authentication bypass

   2. Missing: AMF → UE Security Mode Command
      Impact: High
      Note: Security procedure skipped - potential security risk

   3. Extra Delay: AMF → UE Registration Accept (+2.5s)
      Impact: Low
      Note: Slower than typical (expected: <500ms)
   ```

3. **Completeness Score**
   ```
   Completeness: 75% ●●●●●●●○○○

   Required Messages: 6/8 found (75%)
   Optional Messages: 2/3 found (67%)
   Message Order: Correct ✅
   Timing: Acceptable (some delays)
   ```

---

## Search and Filtering

### Global Search

**Top-right search bar:**

Search across all data by:
- IMSI (e.g., `123456789012345`)
- MSISDN (e.g., `+1234567890`)
- IMEI (e.g., `123456789012345`)
- Transaction ID (e.g., `TID-xxx`)
- IP Address (e.g., `192.168.1.1`)
- Session ID (e.g., `SEID-xxx`)

**Search Results:**
- Grouped by type (Sessions, Subscribers, Events)
- Relevance sorting
- Quick preview
- Click to view details

### Advanced Filtering

**Available on Sessions page:**

**Filter Panel:**

1. **Time Range**
   - Last hour
   - Last 24 hours
   - Last 7 days
   - Custom range (date picker)

2. **Protocols** (Multi-select)
   - [ ] MAP
   - [ ] CAP
   - [ ] INAP
   - [ ] Diameter
   - [ ] GTP
   - [ ] PFCP
   - [ ] HTTP/2
   - [ ] NGAP
   - [ ] S1AP
   - [ ] NAS

3. **Status** (Multi-select)
   - [ ] Success
   - [ ] Failed
   - [ ] Ongoing
   - [ ] Timeout

4. **Identifiers**
   - IMSI contains: `_______`
   - MSISDN contains: `_______`
   - IMEI contains: `_______`

5. **Network Elements**
   - Source IP: `_______`
   - Destination IP: `_______`
   - Source Port: `_______`
   - Destination Port: `_______`

**Apply Filters** button → Results update

**Save Filter** → Name and save for future use

---

## Visualization

### Ladder Diagrams

**Features:**
- Interactive SVG diagrams
- Zoom in/out (mouse wheel)
- Pan (click and drag)
- Hover for message details
- Click for full decode

**Example:**
```
UE        eNB       MME       SGW       PGW       HSS
|          |         |         |         |         |
|-- Attach Request -->|         |         |         |
|          |-- S1AP --|         |         |         |
|          |         |-- Auth ->|         |         |
|          |         |<- Auth --|         |         |
|          |<- Security Mode ---|         |         |
|-- Security Complete>|         |         |         |
|          |         |-- Update Location ->|
|          |         |<- Insert Subscriber Data --|
|          |         |-- Create Session ---------->|
|          |         |<- Create Session Response --|
|<-- Attach Accept ---|         |         |         |
```

**Controls:**
- 🔍 Zoom in/out
- 🖐️ Pan
- 📷 Export as PNG/SVG
- 🖨️ Print
- ⚙️ Settings (show timing, colors, etc.)

### Charts and Graphs

**Dashboard Charts:**

1. **Traffic Rate** (Line chart)
   - Messages per second
   - Last 1 hour (real-time)
   - Protocol breakdown (stacked area)

2. **Protocol Distribution** (Pie chart)
   - Percentage by protocol
   - Message count
   - Click slice to filter

3. **Success Rate** (Gauge)
   - Overall success rate
   - Color-coded (green >95%, yellow 85-95%, red <85%)

4. **Error Trends** (Bar chart)
   - Errors per hour (last 24 hours)
   - Grouped by error type

---

## AI Features

### Pattern Detection

The AI engine automatically detects:

1. **Anomalies**
   - Unusual message sequences
   - Unexpected parameters
   - Timing anomalies

2. **Performance Issues**
   - Slow responses (>threshold)
   - High failure rates
   - Resource bottlenecks

3. **Security Issues**
   - Authentication failures
   - Potential fraud patterns
   - Abnormal subscriber behavior

4. **Protocol Violations**
   - Non-compliant messages
   - Missing mandatory IEs
   - Invalid state transitions

### Automatic Recommendations

For each detected issue, the AI provides:
- **Root cause analysis**
- **Impact assessment**
- **Step-by-step resolution guide**
- **Related knowledge base articles**
- **Similar past issues**

---

## Administration

### User Management

**Location:** Settings → Users

**Create New User:**

1. Click "+ New User"
2. Fill in details:
   - Username
   - Full Name
   - Email
   - Role (Admin/Operator/Viewer)
   - Password (must meet policy)
3. Save

**Role Permissions:**

| Feature | Admin | Operator | Viewer |
|---------|-------|----------|--------|
| View Sessions | ✅ | ✅ | ✅ |
| Export CDRs | ✅ | ✅ | ✅ |
| Delete Sessions | ✅ | ✅ | ❌ |
| Manage Users | ✅ | ❌ | ❌ |
| System Config | ✅ | ❌ | ❌ |
| Start/Stop Service | ✅ | ❌ | ❌ |

### System Control

**Location:** Settings → System

**Actions:**

1. **Service Control**
   - Start/Stop/Restart application
   - Reload configuration (no downtime)
   - View service status

2. **Log Management**
   - View logs in browser
   - Download log files
   - Clear old logs
   - Configure retention

3. **Backup/Restore**
   - Create database backup
   - Restore from backup
   - Schedule automatic backups
   - Export/import configuration

4. **Updates**
   - Check for updates
   - View changelog
   - Install updates (with rollback option)

---

## Keyboard Shortcuts

Global shortcuts:

- `Ctrl+K` or `/` - Global search
- `Ctrl+D` - Go to Dashboard
- `Ctrl+S` - Go to Sessions
- `Ctrl+F` - Open filter panel
- `Esc` - Close modal/panel
- `?` - Show keyboard shortcuts help

---

## Export Options

### Session Export

**Formats:**
- **CSV** - Spreadsheet format
- **JSON** - API-friendly format
- **XML** - Standard interchange
- **PCAP** - Packet capture (raw)
- **PDF** - Human-readable report

**Export Options:**
- Current view (filtered)
- Selected sessions only
- All sessions (with confirmation)
- Custom field selection

### CDR Export

Automatic CDR generation to:
- `/usr/protei/Protei_Monitoring/cdr/{protocol}/`
- Rotated daily
- Format: CSV or custom

---

## Tips and Tricks

1. **Pin Frequent Searches**
   Save commonly used filters as "Saved Searches" for quick access.

2. **Custom Dashboard**
   Customize which widgets appear on your dashboard in Settings.

3. **Keyboard Navigation**
   Use Tab/Shift+Tab to navigate forms and tables faster.

4. **Browser Notifications**
   Enable browser notifications for critical alerts.

5. **Multi-Tab Workflow**
   Open sessions in new tabs (Ctrl+Click) to compare side-by-side.

---

## Support

For web interface questions:
- **Help Icon** (?) in top-right corner
- **Inline Help** - Hover over field labels
- **Video Tutorials** - Settings → Help → Tutorials
- **Support Email**: support@protei.com
