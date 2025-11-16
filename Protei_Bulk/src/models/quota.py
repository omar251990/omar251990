#!/usr/bin/env python3
"""
Quota Usage Models
Tracks message quotas and usage
"""

from sqlalchemy import Column, Integer, BigInteger, String, Date, TIMESTAMP, Numeric, ForeignKey, CheckConstraint, Index, UniqueConstraint
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


class QuotaUsage(Base):
    """Quota Usage Tracking Model"""
    __tablename__ = 'tbl_quota_usage'

    usage_id = Column(BigInteger, primary_key=True, autoincrement=True)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='CASCADE'))

    # Usage Period
    period_date = Column(Date, nullable=False)
    period_type = Column(String(20), default='DAILY')

    # Messages
    messages_sent = Column(Integer, default=0)
    messages_delivered = Column(Integer, default=0)
    messages_failed = Column(Integer, default=0)

    # TPS Tracking
    peak_tps = Column(Numeric(10, 2), default=0.0)
    avg_tps = Column(Numeric(10, 2), default=0.0)

    # Cost
    total_cost = Column(Numeric(15, 4), default=0.0)

    # Timestamps
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Relationships
    user = relationship('User', foreign_keys=[user_id])

    __table_args__ = (
        CheckConstraint("period_type IN ('HOURLY', 'DAILY', 'MONTHLY')", name='check_period_type'),
        UniqueConstraint('user_id', 'period_date', 'period_type', name='uq_quota_user_period'),
        Index('idx_quota_customer', 'customer_id'),
        Index('idx_quota_user', 'user_id'),
        Index('idx_quota_period', 'period_date'),
    )

    def __repr__(self):
        return f"<QuotaUsage(user_id={self.user_id}, date={self.period_date}, sent={self.messages_sent})>"
