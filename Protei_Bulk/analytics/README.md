# Advanced Analytics Engine

Comprehensive analytics and reporting system for Protei_Bulk platform.

## Features

### Real-time Metrics
- **Message Metrics**: TPS, delivery rates, response times, status breakdown
- **Campaign Metrics**: Progress tracking, send rates, completion estimates
- **System Metrics**: CPU, memory, disk usage, connection pools, queue depth
- **Account Metrics**: Usage tracking, credit consumption, TPS utilization
- **Route Metrics**: SMSC performance, success rates, utilization

### Trend Analysis
- Time series data aggregation (minute, hour, day, week, month)
- Trend direction detection (up, down, stable)
- Statistical analysis (min, max, average, standard deviation)
- Configurable time windows

### Predictive Analytics
- Message volume forecasting (24h, 7d ahead)
- Anomaly detection using statistical methods
- Capacity planning recommendations
- Confidence scoring

### Report Generation
- Multiple formats: PDF, Excel, CSV, JSON
- Message delivery reports
- Campaign performance reports
- Account usage and billing reports
- System utilization reports
- Scheduled report generation
- Email delivery

## Architecture

```
analytics/
├── models/
│   └── metrics.py          # Data models for metrics
├── services/
│   ├── analytics_engine.py # Core analytics engine
│   └── report_generator.py # Report generation
└── reports/                # Generated reports output
```

## API Endpoints

### Real-time Metrics

#### Get Real-time Message Metrics
```http
GET /api/v1/analytics/metrics/messages/realtime?window_seconds=60
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "timestamp": "2025-01-15T10:30:00Z",
    "total_messages": 5420,
    "messages_delivered": 5234,
    "messages_failed": 86,
    "messages_pending": 100,
    "messages_per_second": 90.33,
    "delivery_rate_percent": 96.57,
    "avg_delivery_time_seconds": 2.34,
    "p95_delivery_time_seconds": 4.12,
    "p99_delivery_time_seconds": 8.45
  }
}
```

#### Get Campaign Metrics
```http
GET /api/v1/analytics/metrics/campaigns/{campaign_id}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "campaign_id": "CMP_123",
    "campaign_name": "Q1 Promo Campaign",
    "total_recipients": 100000,
    "processed": 75320,
    "remaining": 24680,
    "progress_percentage": 75.32,
    "delivered": 73450,
    "failed": 1870,
    "avg_send_rate_tps": 85.4,
    "estimated_completion": "2025-01-15T14:23:00Z"
  }
}
```

#### Get System Metrics
```http
GET /api/v1/analytics/metrics/system
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "cpu_usage_percent": 45.3,
    "memory_usage_percent": 62.1,
    "disk_usage_percent": 38.7,
    "network_throughput_mbps": 125.4,
    "active_connections": 342,
    "db_connection_pool_size": 20,
    "db_active_connections": 8
  }
}
```

#### Get Account Metrics
```http
GET /api/v1/analytics/metrics/accounts/{account_id}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "account_id": "ACC_456",
    "account_name": "Acme Corp",
    "messages_sent_today": 12450,
    "messages_sent_month": 345678,
    "current_balance": 5432.10,
    "tps_limit": 100,
    "current_tps": 45.2,
    "tps_utilization_percent": 45.2
  }
}
```

### Trend Analysis

#### Get Message Trend
```http
GET /api/v1/analytics/trends/messages?granularity=hour&hours=24
```

**Parameters:**
- `granularity`: minute, hour, day
- `hours`: Number of hours to look back (1-168)

**Response:**
```json
{
  "status": "success",
  "data": {
    "metric_name": "message_volume",
    "granularity": "hour",
    "data_points": [
      {"timestamp": "2025-01-15T09:00:00Z", "value": 5234},
      {"timestamp": "2025-01-15T10:00:00Z", "value": 5891},
      ...
    ],
    "statistics": {
      "min_value": 4521,
      "max_value": 6234,
      "avg_value": 5432,
      "std_dev": 456.3
    },
    "trend": {
      "direction": "up",
      "percentage": 12.5
    }
  }
}
```

### Predictive Analytics

#### Predict Message Volume
```http
GET /api/v1/analytics/predictions/message-volume?hours_ahead=24
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "metric_name": "message_volume",
    "current_value": 5420,
    "predicted_value": 5891,
    "prediction_timestamp": "2025-01-16T10:30:00Z",
    "confidence_percent": 87.5,
    "forecast_24h": [5420, 5523, 5612, ...],
    "forecast_7d": [5420, 5523, ...],
    "anomaly_detection": {
      "is_anomaly": false,
      "anomaly_score": 0.34
    }
  }
}
```

### Report Generation

#### Generate Message Report
```http
POST /api/v1/analytics/reports/messages?start_date=2025-01-01&end_date=2025-01-15&format=excel
```

**Parameters:**
- `start_date`: Report start date
- `end_date`: Report end date
- `format`: csv, json, excel, pdf
- `account_id`: Filter by account (optional)
- `status`: Filter by status (optional)
- `campaign_id`: Filter by campaign (optional)

**Response:**
```json
{
  "status": "success",
  "data": {
    "report_id": "MSG_20250115_103045",
    "summary": {
      "total_messages": 1234567,
      "delivered": 1198765,
      "failed": 35802,
      "delivery_rate": 97.1
    },
    "file_path": "/reports/message_report_20250115_103045.xlsx",
    "format": "excel",
    "generated_at": "2025-01-15T10:30:45Z"
  }
}
```

#### Generate Campaign Report
```http
POST /api/v1/analytics/reports/campaigns/{campaign_id}?format=pdf
```

#### Generate Account Usage Report
```http
POST /api/v1/analytics/reports/accounts/{account_id}/usage?start_date=2025-01-01&end_date=2025-01-31
```

### Dashboard Summary

#### Get Dashboard Summary
```http
GET /api/v1/analytics/dashboard/summary
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "timestamp": "2025-01-15T10:30:00Z",
    "messages": {
      "total_today": 125340,
      "last_hour": 5420,
      "current_tps": 90.33,
      "delivery_rate": 96.57,
      "avg_delivery_time": 2.34
    },
    "campaigns": {
      "active": 12
    },
    "system": {
      "cpu_usage": 45.3,
      "memory_usage": 62.1,
      "disk_usage": 38.7,
      "active_connections": 342
    }
  }
}
```

## Usage Examples

### Python Client

```python
import requests

API_URL = "http://localhost:8080/api/v1"
API_KEY = "your_api_key_here"

headers = {
    "X-API-Key": API_KEY
}

# Get real-time metrics
response = requests.get(
    f"{API_URL}/analytics/metrics/messages/realtime?window_seconds=300",
    headers=headers
)
metrics = response.json()
print(f"Current TPS: {metrics['data']['messages_per_second']}")
print(f"Delivery Rate: {metrics['data']['delivery_rate_percent']}%")

# Get trend analysis
response = requests.get(
    f"{API_URL}/analytics/trends/messages?granularity=hour&hours=24",
    headers=headers
)
trend = response.json()
print(f"Trend Direction: {trend['data']['trend']['direction']}")
print(f"Trend Change: {trend['data']['trend']['percentage']}%")

# Generate report
response = requests.post(
    f"{API_URL}/analytics/reports/messages",
    headers=headers,
    params={
        "start_date": "2025-01-01T00:00:00",
        "end_date": "2025-01-15T23:59:59",
        "format": "excel"
    }
)
report = response.json()
print(f"Report generated: {report['data']['file_path']}")
```

### JavaScript Client

```javascript
const API_URL = 'http://localhost:8080/api/v1';
const API_KEY = 'your_api_key_here';

// Get real-time metrics
async function getRealTimeMetrics() {
  const response = await fetch(
    `${API_URL}/analytics/metrics/messages/realtime?window_seconds=60`,
    {
      headers: {
        'X-API-Key': API_KEY
      }
    }
  );
  const data = await response.json();
  console.log(`Current TPS: ${data.data.messages_per_second}`);
}

// Get campaign progress
async function getCampaignProgress(campaignId) {
  const response = await fetch(
    `${API_URL}/analytics/metrics/campaigns/${campaignId}`,
    {
      headers: {
        'X-API-Key': API_KEY
      }
    }
  );
  const data = await response.json();
  console.log(`Progress: ${data.data.progress_percentage}%`);
}

// Generate report
async function generateReport() {
  const response = await fetch(
    `${API_URL}/analytics/reports/messages?start_date=2025-01-01&end_date=2025-01-15&format=csv`,
    {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY
      }
    }
  );
  const data = await response.json();
  console.log(`Report: ${data.data.file_path}`);
}
```

## Performance Metrics

### Target Performance
- **Metric Collection**: <10ms overhead
- **Real-time Queries**: <100ms response time
- **Trend Analysis**: <500ms for 24h data
- **Report Generation**: <5s for 100K records

### Optimization Tips
1. Use appropriate time windows for real-time metrics
2. Cache frequently accessed metrics
3. Run heavy reports asynchronously
4. Use database indexes on timestamp columns
5. Implement data retention policies

## Monitoring and Alerts

### Built-in Alert Rules
The analytics engine supports configurable alert rules:

```python
from analytics.models.metrics import AlertRule

rule = AlertRule(
    rule_id="high_failure_rate",
    rule_name="High Message Failure Rate",
    metric_name="message_failure_rate",
    condition="gt",
    threshold=5.0,  # 5%
    duration=300,  # 5 minutes
    severity="critical",
    notification_channels=["email", "sms"]
)
```

### Threshold Defaults
- High CPU: >80%
- High Memory: >85%
- High Disk: >90%
- Low Delivery Rate: <95%
- High Failure Rate: >5%

## Data Retention

Analytics data is retained according to the following policy:

| Granularity | Retention Period |
|-------------|------------------|
| Real-time (1min) | 24 hours |
| Hourly | 30 days |
| Daily | 1 year |
| Monthly | 3 years |

## Integration with Web Dashboard

The analytics API seamlessly integrates with the React web dashboard:

```jsx
import { useState, useEffect } from 'react';
import { analyticsAPI } from './services/api';

function Dashboard() {
  const [metrics, setMetrics] = useState(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      const response = await analyticsAPI.dashboardSummary();
      setMetrics(response.data);
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000); // Refresh every 5s

    return () => clearInterval(interval);
  }, []);

  return (
    <div>
      <h1>Dashboard</h1>
      {metrics && (
        <>
          <MetricCard title="TPS" value={metrics.messages.current_tps} />
          <MetricCard title="Delivery Rate" value={`${metrics.messages.delivery_rate}%`} />
          <MetricCard title="Active Campaigns" value={metrics.campaigns.active} />
        </>
      )}
    </div>
  );
}
```

## Advanced Features

### Custom Aggregations
Define custom aggregation queries:

```python
from analytics.services.analytics_engine import AnalyticsEngine

engine = AnalyticsEngine(db_factory)

# Custom query
result = engine.custom_aggregation(
    metric="delivery_time",
    aggregation="percentile",
    percentile=95,
    group_by=["smsc_id", "hour"],
    filters={"date": "2025-01-15"}
)
```

### Real-time Streaming
Subscribe to real-time metric updates via WebSocket:

```javascript
const socket = io('ws://localhost:8080');

socket.on('metrics:realtime', (data) => {
  console.log('Real-time TPS:', data.messages_per_second);
});
```

## Troubleshooting

### High Memory Usage
If analytics consume too much memory:
1. Reduce time series buffer size (default: 1000 data points)
2. Implement data aggregation before storage
3. Use external time series database (InfluxDB, Prometheus)

### Slow Reports
If report generation is slow:
1. Add database indexes on frequently queried columns
2. Use materialized views for common aggregations
3. Generate reports asynchronously
4. Implement pagination for large result sets

## References

- [Metrics Data Models](./models/metrics.py)
- [Analytics Engine](./services/analytics_engine.py)
- [Report Generator](./services/report_generator.py)
- [API Documentation](/api/docs)
