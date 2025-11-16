#!/usr/bin/env python3
"""
Analytics Engine
Real-time analytics, trend analysis, and predictive analytics
"""

import asyncio
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from collections import defaultdict, deque
import statistics
import math
import json

from sqlalchemy import func, and_, or_
from sqlalchemy.orm import Session

from ..models.metrics import (
    MessageMetrics, CampaignMetrics, SystemMetrics, AccountMetrics,
    RouteMetrics, TrendData, PredictiveAnalysis, TimeGranularity,
    AnalyticsConstants
)


class AnalyticsEngine:
    """
    Advanced analytics engine for Protei_Bulk
    Provides real-time metrics, trend analysis, and predictive analytics
    """

    def __init__(self, db_session_factory):
        self.db_factory = db_session_factory

        # In-memory time series storage (for real-time metrics)
        self.message_metrics_ts = deque(maxlen=1000)
        self.system_metrics_ts = deque(maxlen=1000)

        # Aggregation caches
        self.hourly_cache = {}
        self.daily_cache = {}

        # Alert state
        self.active_alerts = {}

    # ========== Real-time Metrics ==========

    def get_realtime_message_metrics(self, db: Session, window_seconds: int = 60) -> MessageMetrics:
        """Get real-time message metrics for the last N seconds"""
        from src.models.message import Message, MessageStatus

        cutoff_time = datetime.utcnow() - timedelta(seconds=window_seconds)

        # Query message statistics
        stats = db.query(
            func.count(Message.id).label('total'),
            func.sum(func.case([(Message.status == MessageStatus.SENT, 1)], else_=0)).label('sent'),
            func.sum(func.case([(Message.status == MessageStatus.DELIVERED, 1)], else_=0)).label('delivered'),
            func.sum(func.case([(Message.status == MessageStatus.FAILED, 1)], else_=0)).label('failed'),
            func.sum(func.case([(Message.status == MessageStatus.PENDING, 1)], else_=0)).label('pending'),
            func.sum(func.case([(Message.status == MessageStatus.REJECTED, 1)], else_=0)).label('rejected')
        ).filter(Message.created_at >= cutoff_time).first()

        # Calculate delivery times
        delivery_times = db.query(
            func.extract('epoch', Message.delivered_at - Message.created_at).label('delivery_time')
        ).filter(
            and_(
                Message.status == MessageStatus.DELIVERED,
                Message.created_at >= cutoff_time,
                Message.delivered_at.isnot(None)
            )
        ).all()

        delivery_times_list = [dt[0] for dt in delivery_times if dt[0] is not None]

        # Calculate percentiles
        avg_delivery = statistics.mean(delivery_times_list) if delivery_times_list else 0
        p95_delivery = self._percentile(delivery_times_list, 95) if len(delivery_times_list) > 5 else 0
        p99_delivery = self._percentile(delivery_times_list, 99) if len(delivery_times_list) > 10 else 0

        # Calculate TPS
        tps = (stats.total or 0) / window_seconds if window_seconds > 0 else 0

        metrics = MessageMetrics(
            timestamp=datetime.utcnow(),
            total_messages=stats.total or 0,
            messages_sent=stats.sent or 0,
            messages_delivered=stats.delivered or 0,
            messages_failed=stats.failed or 0,
            messages_pending=stats.pending or 0,
            messages_rejected=stats.rejected or 0,
            avg_delivery_time=avg_delivery,
            p95_delivery_time=p95_delivery,
            p99_delivery_time=p99_delivery,
            messages_per_second=tps
        )

        # Store in time series
        self.message_metrics_ts.append(metrics)

        return metrics

    def get_campaign_metrics(self, db: Session, campaign_id: str) -> Optional[CampaignMetrics]:
        """Get real-time metrics for a specific campaign"""
        from src.models.campaign import Campaign
        from src.models.message import Message, MessageStatus

        campaign = db.query(Campaign).filter(Campaign.campaign_id == campaign_id).first()
        if not campaign:
            return None

        # Count messages by status
        stats = db.query(
            func.count(Message.id).label('total'),
            func.sum(func.case([(Message.status == MessageStatus.SENT, 1)], else_=0)).label('sent'),
            func.sum(func.case([(Message.status == MessageStatus.DELIVERED, 1)], else_=0)).label('delivered'),
            func.sum(func.case([(Message.status == MessageStatus.FAILED, 1)], else_=0)).label('failed'),
            func.sum(func.case([(Message.status == MessageStatus.PENDING, 1)], else_=0)).label('pending')
        ).filter(Message.campaign_id == campaign_id).first()

        processed = stats.total or 0
        remaining = max(0, campaign.total_recipients - processed)
        progress = (processed / campaign.total_recipients * 100) if campaign.total_recipients > 0 else 0

        # Calculate send rate
        if campaign.started_at:
            elapsed = (datetime.utcnow() - campaign.started_at).total_seconds()
            send_rate = processed / elapsed if elapsed > 0 else 0
        else:
            send_rate = 0

        # Estimate completion
        if send_rate > 0 and remaining > 0:
            est_seconds = remaining / send_rate
            est_completion = datetime.utcnow() + timedelta(seconds=est_seconds)
        else:
            est_completion = None

        return CampaignMetrics(
            campaign_id=campaign_id,
            campaign_name=campaign.name,
            timestamp=datetime.utcnow(),
            total_recipients=campaign.total_recipients,
            processed=processed,
            remaining=remaining,
            progress_percentage=progress,
            sent=stats.sent or 0,
            delivered=stats.delivered or 0,
            failed=stats.failed or 0,
            pending=stats.pending or 0,
            avg_send_rate=send_rate,
            estimated_completion=est_completion,
            actual_completion=campaign.completed_at
        )

    def get_system_metrics(self) -> SystemMetrics:
        """Get current system resource metrics"""
        import psutil

        # CPU and Memory
        cpu_percent = psutil.cpu_percent(interval=1)
        memory = psutil.virtual_memory()
        disk = psutil.disk_usage('/')

        # Network (simplified - would need more detailed implementation)
        network = psutil.net_io_counters()

        metrics = SystemMetrics(
            timestamp=datetime.utcnow(),
            cpu_usage=cpu_percent,
            memory_usage=memory.percent,
            disk_usage=disk.percent,
            network_throughput_mbps=(network.bytes_sent + network.bytes_recv) / 1024 / 1024,
            active_connections=len(psutil.net_connections())
        )

        self.system_metrics_ts.append(metrics)

        return metrics

    def get_account_metrics(self, db: Session, account_id: str) -> Optional[AccountMetrics]:
        """Get usage metrics for a specific account"""
        from src.models.user import Account
        from src.models.message import Message

        account = db.query(Account).filter(Account.account_id == account_id).first()
        if not account:
            return None

        today_start = datetime.utcnow().replace(hour=0, minute=0, second=0, microsecond=0)
        month_start = today_start.replace(day=1)

        # Messages sent today
        msgs_today = db.query(func.count(Message.id)).filter(
            and_(
                Message.account_id == account_id,
                Message.created_at >= today_start
            )
        ).scalar() or 0

        # Messages sent this month
        msgs_month = db.query(func.count(Message.id)).filter(
            and_(
                Message.account_id == account_id,
                Message.created_at >= month_start
            )
        ).scalar() or 0

        # Total messages
        msgs_total = db.query(func.count(Message.id)).filter(
            Message.account_id == account_id
        ).scalar() or 0

        # Calculate current TPS (last minute)
        one_min_ago = datetime.utcnow() - timedelta(minutes=1)
        msgs_last_min = db.query(func.count(Message.id)).filter(
            and_(
                Message.account_id == account_id,
                Message.created_at >= one_min_ago
            )
        ).scalar() or 0
        current_tps = msgs_last_min / 60.0

        tps_utilization = (current_tps / account.tps_limit * 100) if account.tps_limit > 0 else 0

        return AccountMetrics(
            account_id=account_id,
            account_name=account.name,
            timestamp=datetime.utcnow(),
            messages_sent_today=msgs_today,
            messages_sent_month=msgs_month,
            total_messages_alltime=msgs_total,
            current_balance=account.balance,
            tps_limit=account.tps_limit,
            current_tps=current_tps,
            tps_utilization=tps_utilization
        )

    # ========== Trend Analysis ==========

    def get_message_trend(self, db: Session, granularity: TimeGranularity, hours: int = 24) -> TrendData:
        """Get message volume trend over time"""
        from src.models.message import Message

        end_time = datetime.utcnow()
        start_time = end_time - timedelta(hours=hours)

        # Determine time bucket size
        if granularity == TimeGranularity.MINUTE:
            bucket = "1 minute"
            date_trunc = 'minute'
        elif granularity == TimeGranularity.HOUR:
            bucket = "1 hour"
            date_trunc = 'hour'
        else:
            bucket = "1 day"
            date_trunc = 'day'

        # Query time series data
        results = db.query(
            func.date_trunc(date_trunc, Message.created_at).label('time_bucket'),
            func.count(Message.id).label('count')
        ).filter(
            Message.created_at.between(start_time, end_time)
        ).group_by('time_bucket').order_by('time_bucket').all()

        # Build trend data
        trend = TrendData(
            metric_name="message_volume",
            granularity=granularity
        )

        values = []
        for row in results:
            count = row.count
            trend.add_data_point(row.time_bucket, count)
            values.append(count)

        if values:
            trend.min_value = min(values)
            trend.max_value = max(values)
            trend.avg_value = statistics.mean(values)
            trend.std_dev = statistics.stdev(values) if len(values) > 1 else 0

            # Simple trend detection
            if len(values) >= 2:
                first_half_avg = statistics.mean(values[:len(values)//2])
                second_half_avg = statistics.mean(values[len(values)//2:])

                if second_half_avg > first_half_avg * 1.1:
                    trend.trend_direction = "up"
                    trend.trend_percentage = ((second_half_avg - first_half_avg) / first_half_avg * 100)
                elif second_half_avg < first_half_avg * 0.9:
                    trend.trend_direction = "down"
                    trend.trend_percentage = ((first_half_avg - second_half_avg) / first_half_avg * 100)

        return trend

    # ========== Predictive Analytics ==========

    def predict_message_volume(self, db: Session, hours_ahead: int = 24) -> PredictiveAnalysis:
        """Predict message volume using simple linear regression"""
        from src.models.message import Message

        # Get historical data (last 7 days)
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=7)

        results = db.query(
            func.date_trunc('hour', Message.created_at).label('hour'),
            func.count(Message.id).label('count')
        ).filter(
            Message.created_at.between(start_time, end_time)
        ).group_by('hour').order_by('hour').all()

        if len(results) < AnalyticsConstants.MIN_DATA_POINTS_FOR_PREDICTION:
            return PredictiveAnalysis(
                metric_name="message_volume",
                current_value=0,
                predicted_value=0,
                prediction_timestamp=datetime.utcnow() + timedelta(hours=hours_ahead),
                confidence=0
            )

        # Extract values
        values = [row.count for row in results]
        current_value = values[-1] if values else 0

        # Simple moving average prediction
        window_size = min(24, len(values))
        recent_avg = statistics.mean(values[-window_size:])

        # Detect anomalies
        mean_val = statistics.mean(values)
        std_val = statistics.stdev(values) if len(values) > 1 else 0
        is_anomaly = abs(current_value - mean_val) > (AnalyticsConstants.ANOMALY_STD_DEV_MULTIPLIER * std_val)

        # Simple forecast (using recent average)
        forecast_24h = [recent_avg] * 24
        forecast_7d = [recent_avg] * (24 * 7)

        # Confidence based on data stability
        confidence = max(0, min(100, 100 - (std_val / mean_val * 100) if mean_val > 0 else 50))

        return PredictiveAnalysis(
            metric_name="message_volume",
            current_value=current_value,
            predicted_value=recent_avg,
            prediction_timestamp=datetime.utcnow() + timedelta(hours=hours_ahead),
            confidence=confidence,
            forecast_24h=forecast_24h,
            forecast_7d=forecast_7d,
            is_anomaly=is_anomaly,
            anomaly_score=abs(current_value - mean_val) / std_val if std_val > 0 else 0
        )

    # ========== Utility Methods ==========

    @staticmethod
    def _percentile(data: List[float], percentile: int) -> float:
        """Calculate percentile of a dataset"""
        if not data:
            return 0.0
        sorted_data = sorted(data)
        index = (len(sorted_data) - 1) * percentile / 100
        floor = math.floor(index)
        ceil = math.ceil(index)
        if floor == ceil:
            return sorted_data[int(index)]
        d0 = sorted_data[int(floor)] * (ceil - index)
        d1 = sorted_data[int(ceil)] * (index - floor)
        return d0 + d1

    def export_metrics_json(self, metrics: Any) -> str:
        """Export metrics to JSON"""
        from dataclasses import asdict
        return json.dumps(asdict(metrics), default=str, indent=2)
