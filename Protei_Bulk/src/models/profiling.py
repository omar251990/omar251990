#!/usr/bin/env python3
"""
Profiling Models
Subscriber profile, segmentation, and privacy models
"""

from sqlalchemy import Column, Integer, BigInteger, String, Text, Boolean, TIMESTAMP, Date, Numeric, ForeignKey, CheckConstraint, Index, UniqueConstraint
from sqlalchemy.dialects.postgresql import JSONB, INET
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


class AttributeSchema(Base):
    """Attribute Schema Model - Dynamic Field Definitions"""
    __tablename__ = 'tbl_attribute_schema'

    attribute_id = Column(BigInteger, primary_key=True, autoincrement=True)
    attribute_name = Column(String(64), unique=True, nullable=False)
    attribute_code = Column(String(64), unique=True, nullable=False)
    display_name = Column(String(128), nullable=False)
    description = Column(Text)

    # Data Type & Validation
    data_type = Column(String(20), nullable=False)
    allowed_values = Column(JSONB, default=[])
    validation_regex = Column(String(255))
    min_value = Column(Numeric(15, 4))
    max_value = Column(Numeric(15, 4))

    # Field Configuration
    is_required = Column(Boolean, default=False)
    is_searchable = Column(Boolean, default=True)
    is_visible_to_cp = Column(Boolean, default=True)
    is_encrypted = Column(Boolean, default=False)

    # Privacy & Security
    privacy_level = Column(String(20), default='PUBLIC')
    requires_permission = Column(String(100))

    # Display & UI
    display_order = Column(Integer, default=100)
    category = Column(String(64))
    icon = Column(String(64))
    help_text = Column(Text)

    # Status
    is_active = Column(Boolean, default=True)

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Metadata
    metadata = Column(JSONB, default={})

    __table_args__ = (
        CheckConstraint("data_type IN ('STRING', 'INTEGER', 'DECIMAL', 'BOOLEAN', 'ENUM', 'JSON', 'DATE', 'DATETIME')", name='check_attr_data_type'),
        CheckConstraint("privacy_level IN ('PUBLIC', 'SENSITIVE', 'CONFIDENTIAL', 'RESTRICTED')", name='check_attr_privacy_level'),
        Index('idx_attribute_name', 'attribute_name'),
        Index('idx_attribute_code', 'attribute_code'),
        Index('idx_attribute_searchable', 'is_searchable'),
        Index('idx_attribute_active', 'is_active'),
    )

    def __repr__(self):
        return f"<AttributeSchema(id={self.attribute_id}, name='{self.attribute_name}', type='{self.data_type}')>"


class Profile(Base):
    """Profile Model - Subscriber Profiles"""
    __tablename__ = 'tbl_profiles'

    profile_id = Column(BigInteger, primary_key=True, autoincrement=True)

    # Identity (Privacy-Protected)
    msisdn_hash = Column(String(64), unique=True, nullable=False)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))

    # Standard Attributes
    gender = Column(String(20))
    age = Column(Integer)
    date_of_birth = Column(Date)
    language = Column(String(20))

    # Location
    country_code = Column(String(10))
    region = Column(String(64))
    city = Column(String(64))
    postal_code = Column(String(20))

    # Device & Technology
    device_type = Column(String(20))
    device_model = Column(String(128))
    os_version = Column(String(64))

    # Service Info
    plan_type = Column(String(20))
    subscription_date = Column(Date)
    last_recharge_date = Column(Date)

    # Behavioral
    interests = Column(JSONB, default=[])
    preferences = Column(JSONB, default={})

    # Activity
    last_activity_date = Column(Date)
    last_message_sent = Column(Date)
    last_message_received = Column(Date)
    total_messages_sent = Column(Integer, default=0)
    total_messages_received = Column(Integer, default=0)

    # Status
    status = Column(String(20), default='ACTIVE')
    opt_in_marketing = Column(Boolean, default=False)
    opt_in_sms = Column(Boolean, default=True)

    # Custom Attributes (Dynamic)
    custom_attributes = Column(JSONB, default={})

    # Privacy & Consent
    consent_date = Column(TIMESTAMP)
    consent_version = Column(String(20))
    gdpr_compliant = Column(Boolean, default=True)
    data_retention_date = Column(Date)

    # Audit
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())
    imported_at = Column(TIMESTAMP)
    imported_by = Column(String(64))

    # Metadata
    metadata = Column(JSONB, default={})
    tags = Column(JSONB, default=[])

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    group_memberships = relationship('ProfileGroupMember', back_populates='profile')

    __table_args__ = (
        CheckConstraint("gender IN ('MALE', 'FEMALE', 'OTHER', 'UNKNOWN')", name='check_profile_gender'),
        CheckConstraint("device_type IN ('ANDROID', 'IOS', 'FEATURE_PHONE', 'OTHER', 'UNKNOWN')", name='check_profile_device_type'),
        CheckConstraint("plan_type IN ('PREPAID', 'POSTPAID', 'VIP', 'CORPORATE', 'OTHER')", name='check_profile_plan_type'),
        CheckConstraint("status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'DELETED')", name='check_profile_status'),
        Index('idx_profile_hash', 'msisdn_hash'),
        Index('idx_profile_customer', 'customer_id'),
        Index('idx_profile_gender', 'gender'),
        Index('idx_profile_age', 'age'),
        Index('idx_profile_region', 'region'),
        Index('idx_profile_city', 'city'),
        Index('idx_profile_device_type', 'device_type'),
        Index('idx_profile_plan_type', 'plan_type'),
        Index('idx_profile_status', 'status'),
        Index('idx_profile_last_activity', 'last_activity_date'),
        Index('idx_profile_opt_in', 'opt_in_marketing'),
        Index('idx_profile_created', 'created_at'),
        Index('idx_profile_gender_age', 'gender', 'age'),
        Index('idx_profile_region_city', 'region', 'city'),
    )

    def __repr__(self):
        return f"<Profile(id={self.profile_id}, gender='{self.gender}', age={self.age}, status='{self.status}')>"


class ProfileGroup(Base):
    """Profile Group Model - Segments/Audiences"""
    __tablename__ = 'tbl_profile_groups'

    group_id = Column(BigInteger, primary_key=True, autoincrement=True)
    group_code = Column(String(64), unique=True, nullable=False)
    group_name = Column(String(255), nullable=False)
    description = Column(Text)

    # Ownership
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))

    # Filter Definition
    filter_query = Column(JSONB, nullable=False)
    filter_sql = Column(Text)

    # Statistics
    record_count = Column(BigInteger, default=0)
    last_count_updated = Column(TIMESTAMP)

    # Refresh Strategy
    is_dynamic = Column(Boolean, default=True)
    refresh_frequency = Column(String(20))
    last_refreshed = Column(TIMESTAMP)
    next_refresh = Column(TIMESTAMP)

    # Usage Tracking
    total_campaigns_sent = Column(Integer, default=0)
    total_messages_sent = Column(BigInteger, default=0)
    last_used_at = Column(TIMESTAMP)

    # Visibility & Sharing
    visibility = Column(String(20), default='PRIVATE')
    shared_with_users = Column(JSONB, default=[])
    shared_with_customers = Column(JSONB, default=[])

    # Status
    is_active = Column(Boolean, default=True)

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Metadata
    metadata = Column(JSONB, default={})
    tags = Column(JSONB, default=[])

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    user = relationship('User', foreign_keys=[user_id])
    members = relationship('ProfileGroupMember', back_populates='group')

    __table_args__ = (
        CheckConstraint("refresh_frequency IN ('REALTIME', 'HOURLY', 'DAILY', 'WEEKLY', 'MONTHLY', 'MANUAL')", name='check_group_refresh_frequency'),
        CheckConstraint("visibility IN ('PRIVATE', 'SHARED', 'PUBLIC')", name='check_group_visibility'),
        Index('idx_group_code', 'group_code'),
        Index('idx_group_customer', 'customer_id'),
        Index('idx_group_user', 'user_id'),
        Index('idx_group_active', 'is_active'),
        Index('idx_group_visibility', 'visibility'),
        Index('idx_group_last_used', 'last_used_at'),
    )

    def __repr__(self):
        return f"<ProfileGroup(id={self.group_id}, name='{self.group_name}', count={self.record_count})>"


class ProfileGroupMember(Base):
    """Profile Group Member Model - Cached Segment Members"""
    __tablename__ = 'tbl_profile_group_members'

    group_id = Column(BigInteger, ForeignKey('tbl_profile_groups.group_id', ondelete='CASCADE'), primary_key=True)
    profile_id = Column(BigInteger, ForeignKey('tbl_profiles.profile_id', ondelete='CASCADE'), primary_key=True)

    # Timestamps
    added_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    group = relationship('ProfileGroup', back_populates='members')
    profile = relationship('Profile', back_populates='group_memberships')

    __table_args__ = (
        Index('idx_group_members_group', 'group_id'),
        Index('idx_group_members_profile', 'profile_id'),
    )

    def __repr__(self):
        return f"<ProfileGroupMember(group_id={self.group_id}, profile_id={self.profile_id})>"


class ProfileImportJob(Base):
    """Profile Import Job Model - Bulk Import Tracking"""
    __tablename__ = 'tbl_profile_import_jobs'

    job_id = Column(BigInteger, primary_key=True, autoincrement=True)
    job_code = Column(String(64), unique=True, nullable=False)

    # Job Details
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))

    # File Info
    file_name = Column(String(255))
    file_size_bytes = Column(BigInteger)
    file_path = Column(String(500))
    file_type = Column(String(20))

    # Column Mapping
    column_mapping = Column(JSONB)

    # Progress
    status = Column(String(20), default='PENDING')
    total_rows = Column(Integer, default=0)
    rows_processed = Column(Integer, default=0)
    rows_imported = Column(Integer, default=0)
    rows_updated = Column(Integer, default=0)
    rows_failed = Column(Integer, default=0)

    # Timing
    started_at = Column(TIMESTAMP)
    completed_at = Column(TIMESTAMP)
    duration_seconds = Column(Integer)

    # Errors
    error_message = Column(Text)
    error_rows = Column(JSONB, default=[])

    # Options
    update_existing = Column(Boolean, default=True)
    skip_duplicates = Column(Boolean, default=False)
    hash_msisdn = Column(Boolean, default=True)

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())

    # Metadata
    metadata = Column(JSONB, default={})

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    user = relationship('User', foreign_keys=[user_id])

    __table_args__ = (
        CheckConstraint("file_type IN ('CSV', 'EXCEL', 'JSON', 'XML')", name='check_import_file_type'),
        CheckConstraint("status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')", name='check_import_status'),
        Index('idx_import_job_code', 'job_code'),
        Index('idx_import_customer', 'customer_id'),
        Index('idx_import_user', 'user_id'),
        Index('idx_import_status', 'status'),
        Index('idx_import_created', 'created_at'),
    )

    def __repr__(self):
        return f"<ProfileImportJob(id={self.job_id}, code='{self.job_code}', status='{self.status}')>"


class ProfileQueryLog(Base):
    """Profile Query Log Model - Privacy Audit Trail"""
    __tablename__ = 'tbl_profile_query_log'

    query_id = Column(BigInteger, primary_key=True, autoincrement=True)

    # Query Details
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))

    # Query Info
    query_type = Column(String(20))
    filter_query = Column(JSONB)
    result_count = Column(BigInteger)

    # Group Reference
    group_id = Column(BigInteger, ForeignKey('tbl_profile_groups.group_id', ondelete='SET NULL'))
    group_name = Column(String(255))

    # Privacy Compliance
    includes_pii = Column(Boolean, default=False)
    approval_required = Column(Boolean, default=False)
    approved_by = Column(String(64))
    approved_at = Column(TIMESTAMP)

    # Performance
    query_time_ms = Column(Integer)
    cache_hit = Column(Boolean, default=False)

    # Source
    ip_address = Column(INET)
    user_agent = Column(Text)

    # Timestamps
    created_at = Column(TIMESTAMP, default=func.now())

    # Metadata
    metadata = Column(JSONB, default={})

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])
    user = relationship('User', foreign_keys=[user_id])
    group = relationship('ProfileGroup', foreign_keys=[group_id])

    __table_args__ = (
        CheckConstraint("query_type IN ('SEARCH', 'SEGMENT', 'EXPORT', 'COUNT', 'CAMPAIGN')", name='check_query_type'),
        Index('idx_query_log_customer', 'customer_id'),
        Index('idx_query_log_user', 'user_id'),
        Index('idx_query_log_type', 'query_type'),
        Index('idx_query_log_created', 'created_at'),
        Index('idx_query_log_group', 'group_id'),
    )

    def __repr__(self):
        return f"<ProfileQueryLog(id={self.query_id}, type='{self.query_type}', count={self.result_count})>"


class ProfileStatistics(Base):
    """Profile Statistics Model - Aggregated Data"""
    __tablename__ = 'tbl_profile_statistics'

    stat_id = Column(BigInteger, primary_key=True, autoincrement=True)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))

    # Time Period
    period_date = Column(Date, nullable=False)
    period_type = Column(String(20), default='DAILY')

    # Profile Counts
    total_profiles = Column(BigInteger, default=0)
    active_profiles = Column(BigInteger, default=0)
    inactive_profiles = Column(BigInteger, default=0)
    new_profiles = Column(BigInteger, default=0)
    updated_profiles = Column(BigInteger, default=0)

    # Demographics
    male_count = Column(BigInteger, default=0)
    female_count = Column(BigInteger, default=0)
    avg_age = Column(Numeric(5, 2))

    # Device Distribution
    android_count = Column(BigInteger, default=0)
    ios_count = Column(BigInteger, default=0)
    feature_phone_count = Column(BigInteger, default=0)

    # Plan Distribution
    prepaid_count = Column(BigInteger, default=0)
    postpaid_count = Column(BigInteger, default=0)

    # Opt-In Rates
    opt_in_marketing_count = Column(BigInteger, default=0)
    opt_in_sms_count = Column(BigInteger, default=0)

    # Created
    created_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    customer = relationship('Customer', foreign_keys=[customer_id])

    __table_args__ = (
        CheckConstraint("period_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')", name='check_stat_period_type'),
        UniqueConstraint('customer_id', 'period_date', 'period_type', name='uq_profile_stat_period'),
        Index('idx_profile_stats_customer', 'customer_id'),
        Index('idx_profile_stats_period', 'period_date'),
        Index('idx_profile_stats_type', 'period_type'),
    )

    def __repr__(self):
        return f"<ProfileStatistics(customer_id={self.customer_id}, date={self.period_date}, total={self.total_profiles})>"
