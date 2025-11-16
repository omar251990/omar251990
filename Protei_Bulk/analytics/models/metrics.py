#!/usr/bin/env python3
"""
Analytics Data Models
Defines metrics, aggregations, and analytics data structures
"""

from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from enum import Enum


class MetricType(Enum):
    """Types of metrics tracked"""
    COUNTER = "counter"
    GAUGE = "gauge"
    HISTOGRAM = "histogram"
    SUMMARY = "summary"


class TimeGranularity(Enum):
    """Time granularity for aggregations"""
    MINUTE = "minute"
    HOUR = "hour"
    DAY = "day"
    WEEK = "week"
    MONTH = "month"


@dataclass
class MessageMetrics:
    """Real-time message statistics"""
    timestamp: datetime
    total_messages: int = 0
    messages_sent: int = 0
    messages_delivered: int = 0
    messages_failed: int = 0
    messages_pending: int = 0
    messages_rejected: int = 0

    # Performance metrics
    avg_delivery_time: float = 0.0  # seconds
    p95_delivery_time: float = 0.0
    p99_delivery_time: float = 0.0

    # Throughput
    messages_per_second: float = 0.0
    peak_tps: float = 0.0

    # Quality metrics
    delivery_rate: float = 0.0  # percentage
    failure_rate: float = 0.0

    def __post_init__(self):
        """Calculate derived metrics"""
        if self.total_messages > 0:
            self.delivery_rate = (self.messages_delivered / self.total_messages) * 100
            self.failure_rate = (self.messages_failed / self.total_messages) * 100


@dataclass
class CampaignMetrics:
    """Campaign performance metrics"""
    campaign_id: str
    campaign_name: str
    timestamp: datetime

    # Progress
    total_recipients: int = 0
    processed: int = 0
    remaining: int = 0
    progress_percentage: float = 0.0

    # Status breakdown
    sent: int = 0
    delivered: int = 0
    failed: int = 0
    pending: int = 0

    # Performance
    avg_send_rate: float = 0.0  # messages per second
    estimated_completion: Optional[datetime] = None
    actual_completion: Optional[datetime] = None

    # Cost (if applicable)
    total_cost: float = 0.0
    cost_per_message: float = 0.0


@dataclass
class SystemMetrics:
    """System resource and health metrics"""
    timestamp: datetime

    # Resource utilization
    cpu_usage: float = 0.0  # percentage
    memory_usage: float = 0.0  # percentage
    disk_usage: float = 0.0  # percentage

    # Network
    network_throughput_mbps: float = 0.0
    active_connections: int = 0

    # Database
    db_connection_pool_size: int = 0
    db_active_connections: int = 0
    db_query_avg_time: float = 0.0  # milliseconds

    # Redis/Queue
    redis_memory_usage: float = 0.0  # MB
    queue_size: int = 0
    queue_processing_rate: float = 0.0

    # SMPP connections
    smpp_connections_active: int = 0
    smpp_connections_total: int = 0
    smpp_avg_response_time: float = 0.0  # milliseconds


@dataclass
class AccountMetrics:
    """Per-account usage and billing metrics"""
    account_id: str
    account_name: str
    timestamp: datetime

    # Usage
    messages_sent_today: int = 0
    messages_sent_month: int = 0
    total_messages_alltime: int = 0

    # Credits (for prepaid)
    current_balance: float = 0.0
    credits_used_today: float = 0.0
    credits_used_month: float = 0.0

    # Limits
    tps_limit: int = 0
    current_tps: float = 0.0
    tps_utilization: float = 0.0  # percentage

    # Quality
    delivery_rate: float = 0.0
    avg_delivery_time: float = 0.0


@dataclass
class RouteMetrics:
    """SMSC route performance metrics"""
    route_id: str
    smsc_name: str
    timestamp: datetime

    # Volume
    messages_sent: int = 0
    messages_delivered: int = 0
    messages_failed: int = 0

    # Performance
    avg_response_time: float = 0.0  # milliseconds
    success_rate: float = 0.0  # percentage

    # Capacity
    current_tps: float = 0.0
    max_tps: int = 0
    utilization: float = 0.0  # percentage

    # Connection health
    connection_status: str = "UNKNOWN"
    last_error: Optional[str] = None
    uptime: float = 0.0  # hours


@dataclass
class TrendData:
    """Time series trend data"""
    metric_name: str
    granularity: TimeGranularity
    data_points: List[Dict[str, Any]] = field(default_factory=list)

    # Statistical summary
    min_value: float = 0.0
    max_value: float = 0.0
    avg_value: float = 0.0
    std_dev: float = 0.0

    # Trend analysis
    trend_direction: str = "stable"  # up, down, stable
    trend_percentage: float = 0.0  # percentage change

    def add_data_point(self, timestamp: datetime, value: float, metadata: Dict = None):
        """Add a data point to the trend"""
        self.data_points.append({
            "timestamp": timestamp.isoformat(),
            "value": value,
            "metadata": metadata or {}
        })


@dataclass
class PredictiveAnalysis:
    """Predictive analytics results"""
    metric_name: str
    current_value: float
    predicted_value: float
    prediction_timestamp: datetime
    confidence: float  # 0-100

    # Forecasting
    forecast_24h: List[float] = field(default_factory=list)
    forecast_7d: List[float] = field(default_factory=list)

    # Anomaly detection
    is_anomaly: bool = False
    anomaly_score: float = 0.0

    # Capacity planning
    capacity_exhaustion_date: Optional[datetime] = None
    recommended_scaling: Optional[str] = None


@dataclass
class ReportConfiguration:
    """Report generation configuration"""
    report_type: str  # messages, campaigns, accounts, system
    report_name: str

    # Time range
    start_date: datetime
    end_date: datetime

    # Filters
    filters: Dict[str, Any] = field(default_factory=dict)

    # Grouping
    group_by: List[str] = field(default_factory=list)

    # Aggregations
    metrics: List[str] = field(default_factory=list)

    # Output
    format: str = "pdf"  # pdf, excel, csv, json
    include_charts: bool = True
    include_summary: bool = True

    # Scheduling (for recurring reports)
    schedule: Optional[str] = None  # cron expression
    email_recipients: List[str] = field(default_factory=list)


@dataclass
class AggregatedReport:
    """Generated report with aggregated data"""
    report_id: str
    configuration: ReportConfiguration
    generated_at: datetime

    # Summary statistics
    summary: Dict[str, Any] = field(default_factory=dict)

    # Detailed data
    data: List[Dict[str, Any]] = field(default_factory=list)

    # Charts (base64 encoded images or chart config)
    charts: List[Dict[str, Any]] = field(default_factory=list)

    # File path (if saved to disk)
    file_path: Optional[str] = None
    file_size: int = 0  # bytes


@dataclass
class AlertRule:
    """Alert rule configuration"""
    rule_id: str
    rule_name: str
    metric_name: str

    # Condition
    condition: str  # gt, lt, eq, ne, gte, lte
    threshold: float
    duration: int = 60  # seconds - how long condition must persist

    # Actions
    severity: str = "warning"  # info, warning, critical
    notification_channels: List[str] = field(default_factory=list)  # email, sms, webhook

    # State
    is_active: bool = True
    last_triggered: Optional[datetime] = None
    trigger_count: int = 0


@dataclass
class PerformanceBenchmark:
    """Performance benchmark results"""
    benchmark_id: str
    benchmark_name: str
    executed_at: datetime
    duration: float  # seconds

    # Test parameters
    target_tps: int
    test_duration: int  # seconds
    concurrent_users: int

    # Results
    actual_tps: float
    total_requests: int
    successful_requests: int
    failed_requests: int

    # Response times
    avg_response_time: float
    median_response_time: float
    p95_response_time: float
    p99_response_time: float
    max_response_time: float

    # Resource usage during test
    peak_cpu: float
    peak_memory: float
    peak_network_mbps: float

    # Pass/Fail
    passed: bool = False
    pass_criteria: Dict[str, Any] = field(default_factory=dict)
    failure_reasons: List[str] = field(default_factory=list)


class AnalyticsConstants:
    """Constants for analytics"""

    # Percentiles to track
    PERCENTILES = [50, 75, 90, 95, 99]

    # Time windows for aggregation
    WINDOWS = {
        "realtime": 60,  # 1 minute
        "short": 300,  # 5 minutes
        "medium": 3600,  # 1 hour
        "long": 86400,  # 24 hours
    }

    # Thresholds
    HIGH_CPU_THRESHOLD = 80.0
    HIGH_MEMORY_THRESHOLD = 85.0
    HIGH_DISK_THRESHOLD = 90.0
    LOW_DELIVERY_RATE_THRESHOLD = 95.0
    HIGH_FAILURE_RATE_THRESHOLD = 5.0

    # Anomaly detection
    ANOMALY_STD_DEV_MULTIPLIER = 3.0
    MIN_DATA_POINTS_FOR_PREDICTION = 10
