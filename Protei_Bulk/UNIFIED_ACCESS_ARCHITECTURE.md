# Protei_Bulk Unified Access & Submission Architecture

## Overview

Protei_Bulk implements a **Unified Access Architecture** that allows users to submit messages through three different channels using a single user account:

1. **Web Portal** - Browser-based interface with username/password + 2FA
2. **HTTP API Gateway** - RESTful API with API Key or Basic Auth
3. **SMPP Gateway** - SMPP protocol v3.3/v3.4/v5.0 with system_id/password

All three channels share the same user database, authentication system, quota management, and campaign processing engine.

---

## Architecture Components

### 1. Unified User Account

Each user account in `tbl_users` has the following unified access fields:

```sql
-- Authentication
username VARCHAR(100)              -- Used for all channels
password_hash VARCHAR(255)         -- Bcrypt hashed password
api_key VARCHAR(64)                -- For HTTP API authentication
api_key_created_at TIMESTAMP       -- API key creation time
two_factor_enabled BOOLEAN         -- 2FA for web login
two_factor_secret VARCHAR(64)      -- TOTP secret for 2FA

-- Access Control
bind_type VARCHAR(20)              -- 'SMPP', 'HTTP', 'BOTH', 'WEB_ONLY'
can_use_smpp BOOLEAN              -- Allow SMPP binding
can_use_http BOOLEAN              -- Allow HTTP API access
can_use_api_bulk BOOLEAN          -- Allow bulk API operations

-- Quotas & Limits
max_msg_per_day INTEGER           -- Daily message quota (default: 500,000)
max_tps INTEGER                   -- Maximum TPS (transactions per second)

-- Allowed Resources
allowed_smsc JSONB                -- List of allowed SMSC IDs ['SMSC1', 'SMSC2']
allowed_sender_ids JSONB          -- List of allowed sender IDs ['MyCompany', 'MyBrand']
```

### 2. Authentication Flows

#### A. Web Portal Authentication

**Flow**:
1. User navigates to `/login`
2. Enters username and password
3. System validates credentials via `unified_auth_service.authenticate_web()`
4. If 2FA enabled, prompts for OTP code
5. On success, generates JWT token
6. Returns JWT for subsequent API calls

**Implementation**: `src/services/unified_auth_service.py:authenticate_web()`

**Example**:
```python
success, user, jwt_token, error = unified_auth_service.authenticate_web(
    db=db,
    username="ahmad",
    password="Password@123",
    otp_code="123456",  # if 2FA enabled
    ip_address="192.168.1.100"
)

if success:
    # Set JWT token in session
    session['token'] = jwt_token
```

#### B. HTTP API Authentication

**Two Options**:

**Option 1: API Key (Recommended)**
```http
POST /api/sendBulk HTTP/1.1
Host: bulk.protei.com
Authorization: Bearer your-api-key-here
Content-Type: application/json
```

**Option 2: Basic Authentication**
```http
POST /api/sendBulk HTTP/1.1
Host: bulk.protei.com
Authorization: Basic YWhtYWQ6UGFzc3dvcmRAMTIz
Content-Type: application/json
```

**Implementation**:
- `src/services/unified_auth_service.py:authenticate_api_key()`
- `src/services/unified_auth_service.py:authenticate_basic_auth()`

**Generating API Key**:
```python
api_key = unified_auth_service.generate_api_key(db, user_id)
# Returns: "abc123def456...xyz" (48 character secure token)
```

#### C. SMPP Bind Authentication

**Flow**:
1. Client connects to SMPP server (port 2775 or 3550 for TLS)
2. Sends `BIND_TRANSMITTER`, `BIND_RECEIVER`, or `BIND_TRANSCEIVER` PDU
3. System validates `system_id` (username) and `password`
4. Checks `can_use_smpp` permission
5. Creates session in `tbl_smpp_sessions` with unique `session_token`
6. Returns `BIND_TRANSMITTER_RESP` with success status

**Implementation**: `src/services/unified_auth_service.py:authenticate_smpp()`

**Example**:
```python
success, user, session_token, error = unified_auth_service.authenticate_smpp(
    db=db,
    system_id="ahmad",
    password="Password@123",
    bind_type="TRANSCEIVER",
    remote_ip="192.168.1.200",
    remote_port=54321
)

if success:
    # Session created, ready to receive submit_sm PDUs
    smpp_session = SMPPSession(session_token=session_token)
```

---

## 3. Message Submission Channels

### A. Web Portal Submission

**URL**: `https://bulk.protei.com/campaigns/create`

**Flow**:
1. User logs in to web portal
2. Navigates to "Create Campaign"
3. Fills 5-step wizard:
   - Step 1: Channel & Sender ID
   - Step 2: Recipients (upload file, select contact list, or manual entry)
   - Step 3: Message content & encoding
   - Step 4: Schedule (immediate or scheduled)
   - Step 5: Review & submit
4. System creates campaign with `submission_channel='WEB'`
5. Messages queued for processing

**Frontend**: `web/src/pages/Campaigns/CreateCampaign.jsx`

### B. HTTP API Submission

**Endpoint**: `POST /api/sendBulk`

**Authentication**: Bearer token (API Key) or Basic Auth

**Request**:
```json
{
  "sender": "MyCompany",
  "recipients": [
    {
      "msisdn": "962788123456",
      "message": "Hello Ahmad!",
      "variables": {"NAME": "Ahmad"}
    },
    {
      "msisdn": "962779234567",
      "message": "Hello Fatima!",
      "variables": {"NAME": "Fatima"}
    }
  ],
  "message": "Default message template: Hello %NAME%",
  "encoding": "GSM7",
  "priority": "NORMAL",
  "dlr_url": "https://myapp.com/dlr",
  "campaign_name": "Summer Promo",
  "schedule_time": "2025-01-20T10:00:00"
}
```

**Response**:
```json
{
  "status": "SUCCESS",
  "campaign_id": "CAMP-ABC123DEF456",
  "total_recipients": 2,
  "accepted": 2,
  "rejected": 0,
  "message": "Campaign created successfully. 2 messages queued, 0 rejected.",
  "errors": []
}
```

**Implementation**: `src/api/routes/bulk_api.py:send_bulk_messages()`

**Quota Checking**:
- Checks `max_msg_per_day` limit
- Returns HTTP 429 if quota exceeded

**Validation**:
- Sender ID must be in `allowed_sender_ids` list
- Recipients validated (must be digits, min 10 characters)
- Message length validated based on encoding

### C. SMPP Submission

**Protocol**: SMPP v3.3, v3.4, or v5.0

**Port**:
- 2775 (standard SMPP)
- 3550 (SMPP over TLS)

**Flow**:
1. Client binds with `system_id` and `password`
2. Receives `BIND_TRANSMITTER_RESP` with status 0 (success)
3. Sends `SUBMIT_SM` PDU with message details
4. System creates message with `submission_channel='SMPP'`
5. Returns `SUBMIT_SM_RESP` with `message_id`
6. Message queued for processing

**SUBMIT_SM PDU**:
```
source_addr:         "MyCompany"         (sender ID)
destination_addr:    "962788123456"      (recipient)
short_message:       "Hello Ahmad!"      (message text)
data_coding:         0x00                (GSM7)
registered_delivery: 0x01                (request DLR)
```

**Response**: `SUBMIT_SM_RESP` with `message_id`

**Implementation**: (To be implemented in SMPP gateway server)

**TPS Enforcement**:
- Each user has `max_tps` limit
- SMPP server throttles `submit_sm` based on TPS
- Exceeding TPS returns `ESME_RTHROTTLED` error

---

## 4. Unified Campaign Flow

All three channels create campaigns that flow through the same processing pipeline:

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────┐
│   Client    │────>│ Auth Service │────>│  Campaign   │────>│  Message │
│ (Web/API/   │     │              │     │  Builder    │     │  Queue   │
│  SMPP)      │     └──────────────┘     └─────────────┘     └──────────┘
└─────────────┘                                                     │
                                                                    v
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────┐
│   SMSC      │<────│   Router     │<────│   Worker    │<────│  Queue   │
│ Connector   │     │              │     │   Node      │     │ Consumer │
└─────────────┘     └──────────────┘     └─────────────┘     └──────────┘
      │
      v
┌─────────────┐     ┌──────────────┐
│     DLR     │────>│   Customer   │
│  Processor  │     │   Callback   │
└─────────────┘     └──────────────┘
```

**Steps**:

1. **Authentication** - User authenticated via Web/HTTP API/SMPP
2. **Campaign Creation** - Campaign record created in `campaigns` table
3. **Message Generation** - Individual messages created in `messages` table
4. **Queue Insertion** - Messages inserted into `tbl_message_queue`
5. **Queue Processing** - Worker nodes consume queue and process messages
6. **Routing** - Messages routed to appropriate SMSC based on rules
7. **SMSC Submission** - Messages sent to SMSC via SMPP
8. **DLR Processing** - Delivery reports received and processed
9. **Customer Callback** - DLR forwarded to customer's `dlr_callback_url`

---

## 5. Delivery Report (DLR) Handling

### DLR Flow

```
┌─────────┐        ┌──────────┐        ┌─────────────┐        ┌──────────┐
│  SMSC   │───────>│   SMPP   │───────>│     DLR     │───────>│ Database │
│         │  DLR   │ Gateway  │  Parse │   Handler   │ Update │          │
└─────────┘        └──────────┘        └─────────────┘        └──────────┘
                                              │
                                              v
                                        ┌──────────┐
                                        │ Customer │
                                        │ Callback │
                                        └──────────┘
```

### DLR Statuses

| Status     | Meaning                        | Message Status |
|------------|--------------------------------|----------------|
| DELIVRD    | Message delivered successfully | DELIVERED      |
| ACCEPTD    | Message accepted by SMSC       | SENT           |
| ENROUTE    | Message in transit             | SENT           |
| UNDELIV    | Message undelivered            | FAILED         |
| REJECTED   | Message rejected by SMSC       | REJECTED       |
| EXPIRED    | Message validity expired       | EXPIRED        |

### Customer Callback

**Callback URL**: Configured in campaign `dlr_callback_url`

**Method**: HTTP POST

**Payload**:
```json
{
  "message_id": "MSG-ABC123DEF456",
  "msisdn": "962788123456",
  "sender": "MyCompany",
  "status": "DELIVRD",
  "error_code": null,
  "dlr_text": "Message delivered successfully",
  "submitted_at": "2025-01-16T10:00:00Z",
  "delivered_at": "2025-01-16T10:00:05Z",
  "timestamp": "2025-01-16T10:00:06Z"
}
```

**Implementation**: `src/services/dlr_handler.py:process_dlr()`

**Retry Logic**:
- Timeout: 10 seconds
- Max retries: 3
- Exponential backoff

---

## 6. Quota & Rate Limiting

### Daily Quota

**Field**: `users.max_msg_per_day`

**Default**: 500,000 messages/day

**Tracking**: `tbl_quota_usage` table

**Checking**:
```python
usage = unified_auth_service.get_daily_usage(db, user_id)
# Returns:
{
    'messages_sent': 12500,
    'messages_delivered': 12400,
    'messages_failed': 100,
    'max_messages': 500000,
    'remaining': 487500,
    'percentage_used': 2.5
}
```

**Enforcement**:
- Web portal: Shows remaining quota on dashboard
- HTTP API: Returns HTTP 429 if exceeded
- SMPP: Returns `ESME_RSUBMITFAIL` error

### TPS (Transactions Per Second)

**Field**: `users.max_tps`

**Default**: Based on account type

**Enforcement**:
- SMPP gateway tracks TPS per session
- Throttles messages to stay within limit
- Returns `ESME_RTHROTTLED` if exceeded

### SMSC Assignment

**Field**: `users.allowed_smsc`

**Example**: `["SMSC1", "SMSC2", "SMSC3"]`

**Enforcement**:
- Messages routed only to allowed SMSCs
- Routing engine checks `allowed_smsc` list
- Rejects if no allowed SMSC available

### Sender ID Restriction

**Field**: `users.allowed_sender_ids`

**Example**: `["MyCompany", "MyBrand", "9627"]`

**Enforcement**:
- HTTP API validates sender ID before accepting
- SMPP validates `source_addr` field
- Returns error if not in allowed list

---

## 7. Security Features

### 1. TLS/SSL Encryption

**Web Portal**:
- HTTPS with TLS 1.2+
- Certificate from trusted CA
- HSTS enabled

**HTTP API**:
- HTTPS mandatory
- TLS 1.2+ required
- Certificate validation

**SMPP over TLS**:
- Port 3550
- TLS 1.2+ with mutual authentication
- Client certificates supported

### 2. Two-Factor Authentication (2FA)

**Method**: TOTP (Time-based One-Time Password)

**Setup**:
1. User enables 2FA in settings
2. System generates TOTP secret
3. QR code displayed for Google Authenticator / Authy
4. User enters 6-digit code to verify

**Login**:
1. User enters username + password
2. System prompts for 6-digit OTP
3. Validates TOTP code (30-second window)
4. Grants access on success

**Implementation**: `src/services/auth.py:enable_2fa()`

### 3. Rate Limiting

**HTTP API**:
- 60 requests/minute per API key
- 1000 requests/hour per IP address
- Burst allowance: 10 requests

**SMPP**:
- TPS limit per session
- Connection limit per user
- Bind attempts: 5/minute

### 4. Audit Logging

**Table**: `tbl_permission_audit` and `tbl_api_requests`

**Logged Events**:
- Authentication attempts (success/failure)
- Message submissions
- DLR callbacks
- API requests
- Permission changes

**Fields Logged**:
- Timestamp
- User ID
- IP address
- User agent
- Action performed
- Result (success/failure)
- Details/metadata

### 5. IP Whitelisting

**Field**: `users.allowed_ips` (JSONB)

**Example**: `["192.168.1.0/24", "10.0.0.50"]`

**Enforcement**:
- HTTP API checks source IP
- SMPP checks remote IP
- Rejects if not in whitelist

---

## 8. API Endpoints Reference

### Authentication

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/login` | POST | None | Web login (username + password + OTP) |
| `/api/logout` | POST | JWT | Logout and invalidate token |
| `/api/refresh` | POST | JWT | Refresh JWT token |

### Bulk Messaging

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/sendBulk` | POST | API Key / Basic | Submit bulk messages |
| `/api/getDLR` | POST | API Key / Basic | Query delivery reports |
| `/api/getBalance` | GET | API Key / Basic | Get account balance & quota |
| `/api/dlrCallback` | POST | None | Receive DLR from SMSC (internal) |

### User Management

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/users/generate-api-key` | POST | JWT | Generate new API key |
| `/api/v1/users/revoke-api-key` | POST | JWT | Revoke API key |
| `/api/v1/users/enable-2fa` | POST | JWT | Enable 2FA |
| `/api/v1/users/disable-2fa` | POST | JWT | Disable 2FA |

### Monitoring

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/sessions/smpp` | GET | JWT | List active SMPP sessions |
| `/api/v1/sessions/disconnect/{session_id}` | POST | JWT | Force disconnect SMPP session |
| `/api/v1/analytics/usage` | GET | JWT | Get usage statistics |
| `/api/v1/analytics/campaigns/{id}` | GET | JWT | Get campaign statistics |

---

## 9. Database Schema

### Key Tables

**users** - Extended with unified access fields
```sql
api_key VARCHAR(64) UNIQUE
bind_type VARCHAR(20) DEFAULT 'BOTH'
can_use_smpp BOOLEAN DEFAULT TRUE
can_use_http BOOLEAN DEFAULT TRUE
max_msg_per_day INTEGER DEFAULT 500000
allowed_smsc JSONB DEFAULT '[]'
allowed_sender_ids JSONB DEFAULT '[]'
```

**campaigns** - Campaign tracking
```sql
campaign_id VARCHAR(64) UNIQUE
submission_channel VARCHAR(20)  -- WEB, HTTP_API, SMPP, SCHEDULER
sender_id VARCHAR(20)
message_content TEXT
total_recipients INTEGER
status VARCHAR(20)              -- DRAFT, APPROVED, RUNNING, COMPLETED
dlr_callback_url VARCHAR(500)
```

**messages** - Individual messages
```sql
message_id VARCHAR(64) UNIQUE
campaign_id BIGINT
from_addr VARCHAR(20)
to_addr VARCHAR(20)
message_text TEXT
submission_channel VARCHAR(20)
status VARCHAR(20)              -- PENDING, QUEUED, SENT, DELIVERED, FAILED
smpp_msg_id VARCHAR(64)
dlr_status VARCHAR(50)
```

**tbl_smpp_sessions** - Active SMPP connections
```sql
session_id BIGSERIAL PRIMARY KEY
user_id BIGINT
system_id VARCHAR(50)
bind_type VARCHAR(20)           -- TRANSMITTER, RECEIVER, TRANSCEIVER
remote_ip INET
session_token VARCHAR(64) UNIQUE
status VARCHAR(20)              -- BOUND, DISCONNECTED, SUSPENDED
current_tps DECIMAL(10, 2)
messages_sent INTEGER
```

**tbl_api_requests** - HTTP API audit log
```sql
request_id BIGSERIAL PRIMARY KEY
endpoint VARCHAR(255)
method VARCHAR(10)
auth_type VARCHAR(20)           -- API_KEY, BASIC_AUTH, JWT
api_key VARCHAR(64)
ip_address INET
response_status INTEGER
response_time_ms INTEGER
```

**tbl_quota_usage** - Daily usage tracking
```sql
user_id BIGINT
period_date DATE
period_type VARCHAR(20)         -- HOURLY, DAILY, MONTHLY
messages_sent INTEGER
messages_delivered INTEGER
messages_failed INTEGER
peak_tps DECIMAL(10, 2)
```

### SQL Functions

**generate_api_key()** - Generate secure API key
```sql
SELECT generate_api_key();
-- Returns: "abc123def456...xyz789"
```

**check_daily_quota(user_id)** - Check if user under quota
```sql
SELECT check_daily_quota(123);
-- Returns: true/false
```

**update_quota_usage(user_id, customer_id, messages_count)** - Update quota
```sql
SELECT update_quota_usage(123, 456, 100);
```

**get_campaign_stats(campaign_id)** - Get campaign statistics
```sql
SELECT * FROM get_campaign_stats(789);
-- Returns: total, pending, queued, sent, delivered, failed, progress_percentage
```

---

## 10. Implementation Files

### Backend Services

| File | Description |
|------|-------------|
| `src/services/unified_auth_service.py` | Unified authentication for all channels |
| `src/services/dlr_handler.py` | DLR processing and callback forwarding |
| `src/api/routes/bulk_api.py` | HTTP API endpoints for bulk messaging |

### Database Models

| File | Description |
|------|-------------|
| `src/models/campaign.py` | Campaign and Message models |
| `src/models/smpp.py` | SMPP session model |
| `src/models/quota.py` | Quota usage model |

### Database Schema

| File | Description |
|------|-------------|
| `database/unified_access_schema.sql` | Complete schema with tables, indexes, functions, triggers |

### Frontend

| File | Description |
|------|-------------|
| `web/src/pages/Campaigns/CreateCampaign.jsx` | Campaign creation wizard |
| `web/src/pages/Campaigns/CampaignList.jsx` | Campaign monitoring |

---

## 11. Usage Examples

### Example 1: Generate API Key

```python
from src.services.unified_auth_service import unified_auth_service

# Generate API key for user
api_key = unified_auth_service.generate_api_key(db, user_id=123)
print(f"Your API key: {api_key}")
# Output: Your API key: abc123def456ghi789jkl012mno345pqr678stu901vwx234yz
```

### Example 2: Submit Bulk Messages via HTTP API

```bash
curl -X POST https://bulk.protei.com/api/sendBulk \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "MyCompany",
    "recipients": [
      {"msisdn": "962788123456", "message": "Hello Ahmad!"},
      {"msisdn": "962779234567", "message": "Hello Fatima!"}
    ],
    "message": "Default message",
    "encoding": "GSM7",
    "dlr_url": "https://myapp.com/dlr"
  }'
```

**Response**:
```json
{
  "status": "SUCCESS",
  "campaign_id": "CAMP-ABC123DEF456",
  "total_recipients": 2,
  "accepted": 2,
  "rejected": 0,
  "message": "Campaign created successfully. 2 messages queued."
}
```

### Example 3: Query Delivery Reports

```bash
curl -X POST https://bulk.protei.com/api/getDLR \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "campaign_id": "CAMP-ABC123DEF456",
    "limit": 100
  }'
```

**Response**:
```json
{
  "status": "SUCCESS",
  "total": 2,
  "records": [
    {
      "message_id": "MSG-111",
      "msisdn": "962788123456",
      "status": "DELIVERED",
      "dlr_status": "DELIVRD",
      "submitted_at": "2025-01-16T10:00:00Z",
      "delivered_at": "2025-01-16T10:00:05Z",
      "error_code": null
    },
    {
      "message_id": "MSG-222",
      "msisdn": "962779234567",
      "status": "DELIVERED",
      "dlr_status": "DELIVRD",
      "submitted_at": "2025-01-16T10:00:00Z",
      "delivered_at": "2025-01-16T10:00:06Z",
      "error_code": null
    }
  ]
}
```

### Example 4: Check Balance and Quota

```bash
curl https://bulk.protei.com/api/getBalance \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response**:
```json
{
  "status": "SUCCESS",
  "balance": 25000.50,
  "credit_limit": 50000.00,
  "currency": "JOD",
  "messages_sent_today": 12500,
  "messages_remaining_today": 487500,
  "daily_quota": 500000
}
```

### Example 5: SMPP Bind (Pseudo-code)

```python
import smpplib

# Connect to SMPP server
client = smpplib.client.Client('bulk.protei.com', 2775)

# Bind as transceiver
client.bind_transceiver(system_id='ahmad', password='Password@123')

# Submit message
client.send_message(
    source_addr='MyCompany',
    destination_addr='962788123456',
    short_message='Hello from SMPP!',
    registered_delivery=True  # Request DLR
)

# Receive DLR
dlr = client.read_pdu()
print(f"DLR Status: {dlr.message_state}")

# Unbind
client.unbind()
```

---

## 12. Admin Control & Monitoring

### View Active SMPP Sessions

```sql
SELECT
    session_id,
    system_id,
    bind_type,
    remote_ip,
    status,
    current_tps,
    messages_sent,
    bound_at
FROM tbl_smpp_sessions
WHERE status = 'BOUND'
ORDER BY bound_at DESC;
```

### Force Disconnect Session

```python
from src.services.unified_auth_service import unified_auth_service

# Disconnect SMPP session
unified_auth_service.disconnect_smpp(db, session_token='abc123...')
```

### View HTTP API Traffic

```sql
SELECT
    DATE_TRUNC('hour', created_at) AS hour,
    COUNT(*) AS requests,
    AVG(response_time_ms) AS avg_response_time,
    COUNT(*) FILTER (WHERE response_status >= 400) AS errors
FROM tbl_api_requests
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour DESC;
```

### Monitor TPS Per User

```sql
SELECT
    u.username,
    s.current_tps,
    s.messages_sent,
    u.max_tps
FROM tbl_smpp_sessions s
JOIN users u ON s.user_id = u.id
WHERE s.status = 'BOUND'
ORDER BY s.current_tps DESC;
```

---

## 13. Troubleshooting

### Common Issues

**Issue**: "Invalid API key"
- **Cause**: API key not generated or revoked
- **Solution**: Generate new API key via `/api/v1/users/generate-api-key`

**Issue**: "Daily quota exceeded"
- **Cause**: User sent more than `max_msg_per_day` messages
- **Solution**: Wait until next day or increase quota

**Issue**: "SMPP bind failed"
- **Cause**: Incorrect credentials or `can_use_smpp=false`
- **Solution**: Verify credentials and check `users.can_use_smpp`

**Issue**: "Sender ID not allowed"
- **Cause**: Sender not in `allowed_sender_ids` list
- **Solution**: Add sender to allowed list or use different sender

**Issue**: "DLR callback failed"
- **Cause**: Customer callback URL unreachable or timeout
- **Solution**: Verify callback URL is accessible and responds within 10 seconds

---

## 14. Performance Considerations

### Scalability

**Horizontal Scaling**:
- Multiple FastAPI instances behind load balancer
- Multiple SMPP gateway instances
- Redis for session sharing
- PostgreSQL read replicas

**Vertical Scaling**:
- Increase worker processes
- Optimize database indexes
- Connection pooling

### Benchmarks

**HTTP API**:
- Target: 1000 requests/second
- Response time: < 100ms
- Concurrent connections: 10,000+

**SMPP Gateway**:
- Target: 10,000 TPS aggregate
- Per-session TPS: Configurable per user
- Concurrent sessions: 1000+

**Database**:
- Message insert: 50,000/second
- Campaign query: < 50ms
- DLR update: 100,000/second

---

## 15. Future Enhancements

- [ ] SMPP gateway server implementation (Twisted/asyncio)
- [ ] Campaign queue with Kafka/RabbitMQ
- [ ] SMSC connector with failover and load balancing
- [ ] Real-time campaign monitoring dashboard
- [ ] Advanced routing rules engine
- [ ] Multi-language message support
- [ ] Rich media (MMS, WhatsApp) support
- [ ] A/B testing for campaigns
- [ ] Predictive DLR analytics

---

## 16. References

- Database Schema: `database/unified_access_schema.sql`
- Authentication Service: `src/services/unified_auth_service.py`
- HTTP API Routes: `src/api/routes/bulk_api.py`
- DLR Handler: `src/services/dlr_handler.py`
- Multi-Tenant Architecture: `MULTITENANT_ARCHITECTURE.md`
- Main README: `README.md`

---

**Version**: 1.0.0
**Last Updated**: 2025-01-16
**Status**: Core Implementation Complete
**© 2025 Protei Corporation**
