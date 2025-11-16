#!/usr/bin/env python3
"""
Routing Models
SMSC connections and routing rules models
"""

from sqlalchemy import Column, Integer, BigInteger, String, Text, Boolean, TIMESTAMP, Numeric, Time, ForeignKey, CheckConstraint, Index, UniqueConstraint
from sqlalchemy.dialects.postgresql import JSONB, INET
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


class SMSCConnection(Base):
    """SMSC Connection Model"""
    __tablename__ = 'tbl_smsc_connections'

    smsc_id = Column(BigInteger, primary_key=True, autoincrement=True)
    smsc_code = Column(String(64), unique=True, nullable=False)
    smsc_name = Column(String(255), nullable=False)

    # Connection Details
    connection_type = Column(String(20), default='SMPP')
    host = Column(String(255), nullable=False)
    port = Column(Integer, nullable=False)

    # Authentication
    username = Column(String(128))
    password = Column(String(255))
    system_id = Column(String(128))
    system_type = Column(String(64))

    # SMPP Configuration
    bind_type = Column(String(20), default='TRANSCEIVER')
    smpp_version = Column(String(10), default='3.4')
    encoding = Column(String(20), default='GSM7')
    ton = Column(Integer, default=1)
    npi = Column(Integer, default=1)

    # Capacity & Limits
    max_tps = Column(Integer, default=100)
    max_connections = Column(Integer, default=10)
    window_size = Column(Integer, default=10)
    throughput_limit = Column(Integer)

    # Classification
    country_code = Column(String(10))
    region = Column(String(64))
    network_operator = Column(String(128))
    mcc_mnc = Column(JSONB, default=[])

    # Routing Behavior
    is_default_route = Column(Boolean, default=False)
    route_mode = Column(String(20), default='ACTIVE')
    priority = Column(Integer, default=100)

    # Cost
    cost_per_sms = Column(Numeric(10, 4), default=0.0)
    currency = Column(String(10), default='USD')

    # DLR Settings
    dlr_callback_url = Column(String(500))
    supports_dlr = Column(Boolean, default=True)

    # Connection Status
    status = Column(String(20), default='DISCONNECTED')
    last_bind_time = Column(TIMESTAMP)
    last_unbind_time = Column(TIMESTAMP)
    last_heartbeat = Column(TIMESTAMP)
    connection_errors = Column(Integer, default=0)

    # Statistics
    total_messages_sent = Column(BigInteger, default=0)
    total_messages_received = Column(BigInteger, default=0)
    current_tps = Column(Numeric(10, 2), default=0.0)
    avg_response_time_ms = Column(Integer, default=0)
    delivery_rate = Column(Numeric(5, 2), default=0.0)

    # Alerts
    alert_on_disconnect = Column(Boolean, default=True)
    alert_on_high_error_rate = Column(Boolean, default=True)
    error_rate_threshold = Column(Numeric(5, 2), default=5.0)
    alert_emails = Column(JSONB, default=[])

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Metadata
    metadata = Column(JSONB, default={})
    notes = Column(Text)

    # Relationships
    routing_rules = relationship('RoutingRule', foreign_keys='RoutingRule.smsc_id', back_populates='smsc')
    statistics = relationship('SMSCStatistics', back_populates='smsc')

    __table_args__ = (
        CheckConstraint("connection_type IN ('SMPP', 'HTTP', 'CUSTOM')", name='check_smsc_connection_type'),
        CheckConstraint("bind_type IN ('TRANSMITTER', 'RECEIVER', 'TRANSCEIVER')", name='check_smsc_bind_type'),
        CheckConstraint("encoding IN ('GSM7', 'UCS2', 'UTF8', 'ASCII')", name='check_smsc_encoding'),
        CheckConstraint("route_mode IN ('ACTIVE', 'STANDBY', 'DISABLED')", name='check_smsc_route_mode'),
        CheckConstraint("status IN ('CONNECTED', 'DISCONNECTED', 'SUSPENDED', 'MAINTENANCE', 'ERROR')", name='check_smsc_status'),
        Index('idx_smsc_code', 'smsc_code'),
        Index('idx_smsc_status', 'status'),
        Index('idx_smsc_country', 'country_code'),
        Index('idx_smsc_route_mode', 'route_mode'),
        Index('idx_smsc_priority', 'priority'),
    )

    def __repr__(self):
        return f"<SMSCConnection(id={self.smsc_id}, code='{self.smsc_code}', status='{self.status}')>"


class RoutingRule(Base):
    """Routing Rule Model"""
    __tablename__ = 'tbl_routing_rules'

    rule_id = Column(BigInteger, primary_key=True, autoincrement=True)
    rule_code = Column(String(64), unique=True, nullable=False)
    rule_name = Column(String(255), nullable=False)

    # Scope
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))

    # Match Conditions
    condition_type = Column(String(20), nullable=False)
    condition_value = Column(String(500), nullable=False)

    # Additional Filters
    msisdn_prefix = Column(String(20))
    sender_id_pattern = Column(String(100))
    message_type = Column(String(20))
    country_code = Column(String(10))
    mcc_mnc = Column(String(10))
    regex_pattern = Column(String(500))

    # Combined Conditions
    combined_conditions = Column(JSONB, default={})

    # Routing Target
    smsc_id = Column(BigInteger, ForeignKey('tbl_smsc_connections.smsc_id', ondelete='CASCADE'))
    fallback_smsc_id = Column(BigInteger, ForeignKey('tbl_smsc_connections.smsc_id', ondelete='SET NULL'))

    # Priority & Control
    priority = Column(Integer, default=100)
    is_active = Column(Boolean, default=True)

    # Time-Based Routing
    enable_time_based = Column(Boolean, default=False)
    active_hours_start = Column(Time)
    active_hours_end = Column(Time)
    active_days = Column(JSONB, default=[])
    timezone = Column(String(50), default='UTC')

    # Load Balancing
    enable_load_balance = Column(Boolean, default=False)
    load_balance_smsc_ids = Column(JSONB, default=[])
    load_balance_weights = Column(JSONB, default={})

    # Cost Optimization
    enable_cost_routing = Column(Boolean, default=False)
    max_cost_per_sms = Column(Numeric(10, 4))

    # Statistics
    total_messages_routed = Column(BigInteger, default=0)
    total_fallbacks_used = Column(BigInteger, default=0)
    last_used_at = Column(TIMESTAMP)

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Metadata
    metadata = Column(JSONB, default={})
    notes = Column(Text)

    # Relationships
    smsc = relationship('SMSCConnection', foreign_keys=[smsc_id], back_populates='routing_rules')
    fallback_smsc = relationship('SMSCConnection', foreign_keys=[fallback_smsc_id])
    routing_logs = relationship('RoutingLog', back_populates='rule')

    __table_args__ = (
        CheckConstraint("condition_type IN ('PREFIX', 'SENDER', 'CUSTOMER', 'COUNTRY', 'MESSAGE_TYPE', 'REGEX', 'COMBINED')", name='check_routing_condition_type'),
        Index('idx_routing_rule_code', 'rule_code'),
        Index('idx_routing_customer', 'customer_id'),
        Index('idx_routing_user', 'user_id'),
        Index('idx_routing_condition_type', 'condition_type'),
        Index('idx_routing_priority', 'priority'),
        Index('idx_routing_active', 'is_active'),
        Index('idx_routing_smsc', 'smsc_id'),
        Index('idx_routing_prefix', 'msisdn_prefix'),
    )

    def __repr__(self):
        return f"<RoutingRule(id={self.rule_id}, name='{self.rule_name}', priority={self.priority})>"


class RoutingLog(Base):
    """Routing Log Model"""
    __tablename__ = 'tbl_routing_logs'

    log_id = Column(BigInteger, primary_key=True, autoincrement=True)

    # Message Reference
    message_id = Column(BigInteger, ForeignKey('messages.id', ondelete='CASCADE'))
    campaign_id = Column(BigInteger, ForeignKey('campaigns.id', ondelete='CASCADE'))

    # Routing Decision
    msisdn = Column(String(20), nullable=False)
    sender_id = Column(String(20))
    message_type = Column(String(20))

    # Rule Applied
    rule_id = Column(BigInteger, ForeignKey('tbl_routing_rules.rule_id', ondelete='SET NULL'))
    rule_name = Column(String(255))

    # SMSC Selection
    selected_smsc_id = Column(BigInteger, ForeignKey('tbl_smsc_connections.smsc_id', ondelete='SET NULL'))
    smsc_name = Column(String(255))

    # Fallback Info
    is_fallback = Column(Boolean, default=False)
    fallback_reason = Column(String(255))
    original_smsc_id = Column(BigInteger)

    # Result
    routing_status = Column(String(20))
    routing_time_ms = Column(Integer)

    # Timestamps
    created_at = Column(TIMESTAMP, default=func.now())

    # Metadata
    metadata = Column(JSONB, default={})

    # Relationships
    rule = relationship('RoutingRule', back_populates='routing_logs')

    __table_args__ = (
        CheckConstraint("routing_status IN ('SUCCESS', 'FALLBACK', 'FAILED', 'NO_ROUTE')", name='check_routing_status'),
        Index('idx_routing_log_message', 'message_id'),
        Index('idx_routing_log_campaign', 'campaign_id'),
        Index('idx_routing_log_rule', 'rule_id'),
        Index('idx_routing_log_smsc', 'selected_smsc_id'),
        Index('idx_routing_log_created', 'created_at'),
        Index('idx_routing_log_msisdn', 'msisdn'),
    )

    def __repr__(self):
        return f"<RoutingLog(id={self.log_id}, msisdn='{self.msisdn}', status='{self.routing_status}')>"


class CountryCode(Base):
    """Country Code Model"""
    __tablename__ = 'tbl_country_codes'

    country_id = Column(BigInteger, primary_key=True, autoincrement=True)
    country_code = Column(String(10), unique=True, nullable=False)
    country_name = Column(String(128), nullable=False)
    iso_code_2 = Column(String(2))
    iso_code_3 = Column(String(3))

    # Mobile Prefixes
    mobile_prefixes = Column(JSONB, default=[])

    # MCC-MNC Codes
    mcc = Column(String(5))
    mnc_list = Column(JSONB, default=[])

    # Operators
    operators = Column(JSONB, default=[])

    # Routing Info
    default_smsc_id = Column(BigInteger, ForeignKey('tbl_smsc_connections.smsc_id', ondelete='SET NULL'))
    region = Column(String(64))

    # Metadata
    is_active = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Relationships
    default_smsc = relationship('SMSCConnection', foreign_keys=[default_smsc_id])

    __table_args__ = (
        Index('idx_country_code', 'country_code'),
        Index('idx_country_iso2', 'iso_code_2'),
        Index('idx_country_region', 'region'),
    )

    def __repr__(self):
        return f"<CountryCode(code='{self.country_code}', name='{self.country_name}')>"


class SMSCStatistics(Base):
    """SMSC Statistics Model"""
    __tablename__ = 'tbl_smsc_statistics'

    stat_id = Column(BigInteger, primary_key=True, autoincrement=True)
    smsc_id = Column(BigInteger, ForeignKey('tbl_smsc_connections.smsc_id', ondelete='CASCADE'))

    # Time Period
    period_start = Column(TIMESTAMP, nullable=False)
    period_end = Column(TIMESTAMP, nullable=False)
    period_type = Column(String(20), default='HOURLY')

    # Traffic
    messages_submitted = Column(Integer, default=0)
    messages_delivered = Column(Integer, default=0)
    messages_failed = Column(Integer, default=0)
    messages_pending = Column(Integer, default=0)

    # Performance
    avg_tps = Column(Numeric(10, 2), default=0.0)
    peak_tps = Column(Numeric(10, 2), default=0.0)
    avg_response_time_ms = Column(Integer, default=0)

    # Delivery
    delivery_rate = Column(Numeric(5, 2), default=0.0)
    error_rate = Column(Numeric(5, 2), default=0.0)

    # Errors
    bind_errors = Column(Integer, default=0)
    submit_errors = Column(Integer, default=0)
    timeout_errors = Column(Integer, default=0)

    # Cost
    total_cost = Column(Numeric(15, 4), default=0.0)

    # Created
    created_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    smsc = relationship('SMSCConnection', back_populates='statistics')

    __table_args__ = (
        CheckConstraint("period_type IN ('MINUTE', 'HOURLY', 'DAILY', 'MONTHLY')", name='check_period_type'),
        UniqueConstraint('smsc_id', 'period_start', 'period_type', name='uq_smsc_stat_period'),
        Index('idx_smsc_stats_smsc', 'smsc_id'),
        Index('idx_smsc_stats_period', 'period_start'),
        Index('idx_smsc_stats_type', 'period_type'),
    )

    def __repr__(self):
        return f"<SMSCStatistics(smsc_id={self.smsc_id}, period={self.period_start})>"
