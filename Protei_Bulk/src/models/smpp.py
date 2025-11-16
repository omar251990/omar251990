#!/usr/bin/env python3
"""
SMPP Models
SMPP session and message models
"""

from sqlalchemy import Column, Integer, BigInteger, String, Text, Boolean, TIMESTAMP, Numeric, ForeignKey, CheckConstraint, Index
from sqlalchemy.dialects.postgresql import JSONB, INET
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


class SMPPSession(Base):
    """SMPP Session Model"""
    __tablename__ = 'tbl_smpp_sessions'

    session_id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='CASCADE'))
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))

    # Connection Details
    system_id = Column(String(50), nullable=False)
    bind_type = Column(String(20), nullable=False)
    remote_ip = Column(INET, nullable=False)
    remote_port = Column(Integer)

    # Session Info
    session_token = Column(String(64), unique=True, nullable=False)
    smpp_version = Column(String(10), default='3.4')

    # Status
    status = Column(String(20), default='BOUND')

    # Throughput
    current_tps = Column(Numeric(10, 2), default=0.0)
    messages_sent = Column(Integer, default=0)
    messages_received = Column(Integer, default=0)

    # Timestamps
    bound_at = Column(TIMESTAMP, default=func.now())
    last_activity_at = Column(TIMESTAMP, default=func.now())
    disconnected_at = Column(TIMESTAMP)

    # Metadata
    metadata = Column(JSONB, default={})

    # Relationships
    user = relationship('User', foreign_keys=[user_id])

    __table_args__ = (
        CheckConstraint("bind_type IN ('TRANSMITTER', 'RECEIVER', 'TRANSCEIVER')", name='check_bind_type'),
        CheckConstraint("status IN ('BOUND', 'DISCONNECTED', 'SUSPENDED')", name='check_session_status'),
        Index('idx_smpp_sessions_user', 'user_id'),
        Index('idx_smpp_sessions_customer', 'customer_id'),
        Index('idx_smpp_sessions_status', 'status'),
        Index('idx_smpp_sessions_token', 'session_token'),
    )

    def __repr__(self):
        return f"<SMPPSession(id={self.session_id}, system_id='{self.system_id}', status='{self.status}')>"
