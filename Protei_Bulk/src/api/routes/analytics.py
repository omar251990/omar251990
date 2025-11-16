#!/usr/bin/env python3
"""
Analytics API Endpoints
Provides real-time metrics, trends, and reports
"""

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import Optional, List
from datetime import datetime, timedelta

from src.core.database import get_db
from src.services.auth import get_current_user
from src.models.user import User
from analytics.services.analytics_engine import AnalyticsEngine
from analytics.services.report_generator import ReportGenerator
from analytics.models.metrics import TimeGranularity

router = APIRouter(prefix="/analytics", tags=["analytics"])


# Initialize analytics services
analytics_engine = None
report_generator = None


def get_analytics_engine(db: Session = Depends(get_db)) -> AnalyticsEngine:
    """Get analytics engine instance"""
    global analytics_engine
    if analytics_engine is None:
        from src.core.database import get_session_factory
        analytics_engine = AnalyticsEngine(get_session_factory())
    return analytics_engine


def get_report_generator(db: Session = Depends(get_db)) -> ReportGenerator:
    """Get report generator instance"""
    global report_generator
    if report_generator is None:
        from src.core.database import get_session_factory
        report_generator = ReportGenerator(get_session_factory())
    return report_generator


# ========== Real-time Metrics Endpoints ==========

@router.get("/metrics/messages/realtime")
async def get_realtime_message_metrics(
    window_seconds: int = Query(60, ge=10, le=3600),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """
    Get real-time message metrics

    - **window_seconds**: Time window in seconds (10-3600)
    """
    metrics = engine.get_realtime_message_metrics(db, window_seconds)
    return {
        "status": "success",
        "data": {
            "timestamp": metrics.timestamp.isoformat(),
            "total_messages": metrics.total_messages,
            "messages_sent": metrics.messages_sent,
            "messages_delivered": metrics.messages_delivered,
            "messages_failed": metrics.messages_failed,
            "messages_pending": metrics.messages_pending,
            "messages_rejected": metrics.messages_rejected,
            "avg_delivery_time_seconds": metrics.avg_delivery_time,
            "p95_delivery_time_seconds": metrics.p95_delivery_time,
            "p99_delivery_time_seconds": metrics.p99_delivery_time,
            "messages_per_second": metrics.messages_per_second,
            "delivery_rate_percent": metrics.delivery_rate,
            "failure_rate_percent": metrics.failure_rate
        }
    }


@router.get("/metrics/campaigns/{campaign_id}")
async def get_campaign_metrics(
    campaign_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """Get real-time metrics for a specific campaign"""
    metrics = engine.get_campaign_metrics(db, campaign_id)

    if not metrics:
        raise HTTPException(status_code=404, detail="Campaign not found")

    return {
        "status": "success",
        "data": {
            "campaign_id": metrics.campaign_id,
            "campaign_name": metrics.campaign_name,
            "timestamp": metrics.timestamp.isoformat(),
            "total_recipients": metrics.total_recipients,
            "processed": metrics.processed,
            "remaining": metrics.remaining,
            "progress_percentage": metrics.progress_percentage,
            "sent": metrics.sent,
            "delivered": metrics.delivered,
            "failed": metrics.failed,
            "pending": metrics.pending,
            "avg_send_rate_tps": metrics.avg_send_rate,
            "estimated_completion": metrics.estimated_completion.isoformat() if metrics.estimated_completion else None,
            "actual_completion": metrics.actual_completion.isoformat() if metrics.actual_completion else None
        }
    }


@router.get("/metrics/system")
async def get_system_metrics(
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """Get current system resource metrics"""
    metrics = engine.get_system_metrics()

    return {
        "status": "success",
        "data": {
            "timestamp": metrics.timestamp.isoformat(),
            "cpu_usage_percent": metrics.cpu_usage,
            "memory_usage_percent": metrics.memory_usage,
            "disk_usage_percent": metrics.disk_usage,
            "network_throughput_mbps": metrics.network_throughput_mbps,
            "active_connections": metrics.active_connections,
            "db_connection_pool_size": metrics.db_connection_pool_size,
            "db_active_connections": metrics.db_active_connections
        }
    }


@router.get("/metrics/accounts/{account_id}")
async def get_account_metrics(
    account_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """Get usage metrics for a specific account"""
    metrics = engine.get_account_metrics(db, account_id)

    if not metrics:
        raise HTTPException(status_code=404, detail="Account not found")

    return {
        "status": "success",
        "data": {
            "account_id": metrics.account_id,
            "account_name": metrics.account_name,
            "timestamp": metrics.timestamp.isoformat(),
            "messages_sent_today": metrics.messages_sent_today,
            "messages_sent_month": metrics.messages_sent_month,
            "total_messages_alltime": metrics.total_messages_alltime,
            "current_balance": metrics.current_balance,
            "credits_used_today": metrics.credits_used_today,
            "credits_used_month": metrics.credits_used_month,
            "tps_limit": metrics.tps_limit,
            "current_tps": metrics.current_tps,
            "tps_utilization_percent": metrics.tps_utilization
        }
    }


# ========== Trend Analysis Endpoints ==========

@router.get("/trends/messages")
async def get_message_trend(
    granularity: str = Query("hour", regex="^(minute|hour|day)$"),
    hours: int = Query(24, ge=1, le=168),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """
    Get message volume trend over time

    - **granularity**: minute, hour, or day
    - **hours**: Number of hours to look back (1-168)
    """
    granularity_enum = TimeGranularity[granularity.upper()]
    trend = engine.get_message_trend(db, granularity_enum, hours)

    return {
        "status": "success",
        "data": {
            "metric_name": trend.metric_name,
            "granularity": trend.granularity.value,
            "data_points": trend.data_points,
            "statistics": {
                "min_value": trend.min_value,
                "max_value": trend.max_value,
                "avg_value": trend.avg_value,
                "std_dev": trend.std_dev
            },
            "trend": {
                "direction": trend.trend_direction,
                "percentage": trend.trend_percentage
            }
        }
    }


# ========== Predictive Analytics Endpoints ==========

@router.get("/predictions/message-volume")
async def predict_message_volume(
    hours_ahead: int = Query(24, ge=1, le=168),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """
    Predict future message volume

    - **hours_ahead**: Hours to predict ahead (1-168)
    """
    prediction = engine.predict_message_volume(db, hours_ahead)

    return {
        "status": "success",
        "data": {
            "metric_name": prediction.metric_name,
            "current_value": prediction.current_value,
            "predicted_value": prediction.predicted_value,
            "prediction_timestamp": prediction.prediction_timestamp.isoformat(),
            "confidence_percent": prediction.confidence,
            "forecast_24h": prediction.forecast_24h,
            "forecast_7d": prediction.forecast_7d,
            "anomaly_detection": {
                "is_anomaly": prediction.is_anomaly,
                "anomaly_score": prediction.anomaly_score
            }
        }
    }


# ========== Report Generation Endpoints ==========

@router.post("/reports/messages")
async def generate_message_report(
    start_date: datetime = Query(..., description="Start date for report"),
    end_date: datetime = Query(..., description="End date for report"),
    format: str = Query("csv", regex="^(csv|json|excel|pdf)$"),
    account_id: Optional[str] = None,
    status: Optional[str] = None,
    campaign_id: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    generator: ReportGenerator = Depends(get_report_generator)
):
    """
    Generate message delivery report

    - **start_date**: Report start date
    - **end_date**: Report end date
    - **format**: csv, json, excel, or pdf
    - **account_id**: Filter by account (optional)
    - **status**: Filter by status (optional)
    - **campaign_id**: Filter by campaign (optional)
    """
    filters = {}
    if account_id:
        filters['account_id'] = account_id
    if status:
        filters['status'] = status
    if campaign_id:
        filters['campaign_id'] = campaign_id

    report = generator.generate_message_report(db, start_date, end_date, format, filters)

    return {
        "status": "success",
        "data": report
    }


@router.post("/reports/campaigns/{campaign_id}")
async def generate_campaign_report(
    campaign_id: str,
    format: str = Query("json", regex="^(csv|json|excel|pdf)$"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    generator: ReportGenerator = Depends(get_report_generator)
):
    """Generate detailed campaign performance report"""
    report = generator.generate_campaign_report(db, campaign_id, format)

    if "error" in report:
        raise HTTPException(status_code=404, detail=report["error"])

    return {
        "status": "success",
        "data": report
    }


@router.post("/reports/accounts/{account_id}/usage")
async def generate_account_usage_report(
    account_id: str,
    start_date: datetime = Query(..., description="Start date for report"),
    end_date: datetime = Query(..., description="End date for report"),
    format: str = Query("excel", regex="^(csv|json|excel|pdf)$"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    generator: ReportGenerator = Depends(get_report_generator)
):
    """Generate account usage and billing report"""
    report = generator.generate_account_usage_report(db, account_id, start_date, end_date, format)

    if "error" in report:
        raise HTTPException(status_code=404, detail=report["error"])

    return {
        "status": "success",
        "data": report
    }


# ========== Dashboard Summary Endpoint ==========

@router.get("/dashboard/summary")
async def get_dashboard_summary(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
    engine: AnalyticsEngine = Depends(get_analytics_engine)
):
    """Get comprehensive dashboard summary with all key metrics"""

    # Get all metrics
    message_metrics = engine.get_realtime_message_metrics(db, 3600)  # Last hour
    system_metrics = engine.get_system_metrics()

    # Get today's totals
    from src.models.message import Message
    from datetime import datetime

    today_start = datetime.utcnow().replace(hour=0, minute=0, second=0, microsecond=0)

    total_today = db.query(func.count(Message.id)).filter(
        Message.created_at >= today_start
    ).scalar() or 0

    # Active campaigns
    from src.models.campaign import Campaign, CampaignStatus

    active_campaigns = db.query(func.count(Campaign.id)).filter(
        Campaign.status.in_([CampaignStatus.RUNNING, CampaignStatus.SCHEDULED])
    ).scalar() or 0

    return {
        "status": "success",
        "data": {
            "timestamp": datetime.utcnow().isoformat(),
            "messages": {
                "total_today": total_today,
                "last_hour": message_metrics.total_messages,
                "current_tps": message_metrics.messages_per_second,
                "delivery_rate": message_metrics.delivery_rate,
                "avg_delivery_time": message_metrics.avg_delivery_time
            },
            "campaigns": {
                "active": active_campaigns
            },
            "system": {
                "cpu_usage": system_metrics.cpu_usage,
                "memory_usage": system_metrics.memory_usage,
                "disk_usage": system_metrics.disk_usage,
                "active_connections": system_metrics.active_connections
            }
        }
    }
