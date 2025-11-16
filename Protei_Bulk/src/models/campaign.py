#!/usr/bin/env python3
"""
Campaign Models
Campaign and message models for unified access
"""

from sqlalchemy import Column, Integer, BigInteger, String, Text, Boolean, TIMESTAMP, Numeric, ForeignKey, CheckConstraint, Index
from sqlalchemy.dialects.postgresql import JSONB, INET
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


class Campaign(Base):
    """Campaign Model"""
    __tablename__ = 'campaigns'

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    campaign_id = Column(String(64), unique=True, nullable=False)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))

    # Campaign Details
    name = Column(String(255), nullable=False)
    description = Column(Text)

    # Submission Channel
    submission_channel = Column(String(20), default='WEB')
    submission_ip = Column(INET)

    # Message Content
    sender_id = Column(String(20), nullable=False)
    message_content = Column(Text, nullable=False)
    encoding = Column(String(20), default='GSM7')
    message_class = Column(String(10), default='FLASH')

    # Recipients
    total_recipients = Column(Integer, nullable=False)
    recipients_source = Column(String(50))  # FILE, CONTACT_LIST, PROFILE, MANUAL, API

    # Status
    status = Column(String(20), default='DRAFT')

    # Schedule
    schedule_type = Column(String(20), default='IMMEDIATE')
    scheduled_time = Column(TIMESTAMP)
    started_at = Column(TIMESTAMP)
    completed_at = Column(TIMESTAMP)

    # Limits
    max_messages_per_day = Column(Integer)
    priority = Column(String(20), default='NORMAL')

    # DLR Settings
    dlr_required = Column(Boolean, default=True)
    dlr_callback_url = Column(String(500))

    # Maker-Checker
    created_by = Column(BigInteger, ForeignKey('users.id'))
    approved_by = Column(BigInteger, ForeignKey('users.id'))
    approved_at = Column(TIMESTAMP)

    # Audit
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Metadata
    metadata = Column(JSONB, default={})

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    user = relationship('User', foreign_keys=[user_id])
    messages = relationship('Message', back_populates='campaign')

    __table_args__ = (
        CheckConstraint("submission_channel IN ('WEB', 'HTTP_API', 'SMPP', 'SCHEDULER')", name='check_submission_channel'),
        CheckConstraint("encoding IN ('GSM7', 'UCS2', 'ASCII')", name='check_encoding'),
        CheckConstraint("status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'SCHEDULED', 'RUNNING', 'PAUSED', 'COMPLETED', 'FAILED', 'CANCELLED')", name='check_campaign_status'),
        CheckConstraint("schedule_type IN ('IMMEDIATE', 'SCHEDULED', 'RECURRING')", name='check_schedule_type'),
        CheckConstraint("priority IN ('CRITICAL', 'HIGH', 'NORMAL', 'LOW')", name='check_priority'),
        Index('idx_campaigns_customer', 'customer_id'),
        Index('idx_campaigns_user', 'user_id'),
        Index('idx_campaigns_status', 'status'),
        Index('idx_campaigns_scheduled', 'scheduled_time'),
        Index('idx_campaigns_created', 'created_at'),
    )

    def __repr__(self):
        return f"<Campaign(id={self.id}, name='{self.name}', status='{self.status}')>"


class Message(Base):
    """Message Model"""
    __tablename__ = 'messages'

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    message_id = Column(String(64), unique=True, nullable=False)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))
    campaign_id = Column(BigInteger, ForeignKey('campaigns.id', ondelete='CASCADE'))

    # Message Details
    from_addr = Column(String(20), nullable=False)
    to_addr = Column(String(20), nullable=False)
    message_text = Column(Text)
    encoding = Column(String(20), default='GSM7')

    # Submission
    submission_channel = Column(String(20), default='WEB')
    submission_timestamp = Column(TIMESTAMP, default=func.now())
    submission_ip = Column(INET)

    # SMPP Specific
    smpp_msg_id = Column(String(64))
    smpp_system_id = Column(String(50))
    esm_class = Column(Integer)
    protocol_id = Column(Integer)
    priority_flag = Column(Integer)
    data_coding = Column(Integer)

    # Status
    status = Column(String(20), default='PENDING')

    # Routing
    smsc_id = Column(String(50))
    route_id = Column(String(50))

    # Timestamps
    queued_at = Column(TIMESTAMP)
    sent_at = Column(TIMESTAMP)
    delivered_at = Column(TIMESTAMP)

    # DLR
    dlr_status = Column(String(50))
    dlr_timestamp = Column(TIMESTAMP)
    dlr_text = Column(Text)
    error_code = Column(String(20))

    # Billing
    cost = Column(Numeric(10, 4), default=0.0)
    parts = Column(Integer, default=1)

    # Metadata
    metadata = Column(JSONB, default={})
    created_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    user = relationship('User', foreign_keys=[user_id])
    campaign = relationship('Campaign', back_populates='messages')

    __table_args__ = (
        CheckConstraint("status IN ('PENDING', 'QUEUED', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED', 'EXPIRED')", name='check_message_status'),
        Index('idx_messages_customer', 'customer_id'),
        Index('idx_messages_user', 'user_id'),
        Index('idx_messages_campaign', 'campaign_id'),
        Index('idx_messages_status', 'status'),
        Index('idx_messages_to_addr', 'to_addr'),
        Index('idx_messages_created', 'submission_timestamp'),
        Index('idx_messages_smpp_msg_id', 'smpp_msg_id'),
    )

    def __repr__(self):
        return f"<Message(id={self.id}, to='{self.to_addr}', status='{self.status}')>"
