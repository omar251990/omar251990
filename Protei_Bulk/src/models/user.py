"""
User, Account, and Role Models
"""

from sqlalchemy import Column, Integer, String, Boolean, DateTime, ForeignKey, ARRAY, Numeric, Text
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import JSONB
import datetime

from src.core.database import Base


class AccountType(Base):
    """Account type lookup table"""
    __tablename__ = "account_types"

    id = Column(Integer, primary_key=True)
    name = Column(String(50), unique=True, nullable=False)
    description = Column(Text)
    hierarchy_level = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)


class Account(Base):
    """Account model - represents companies/resellers"""
    __tablename__ = "accounts"

    id = Column(Integer, primary_key=True)
    account_id = Column(String(64), unique=True, nullable=False)
    account_type_id = Column(Integer, ForeignKey("account_types.id"))
    parent_account_id = Column(Integer, ForeignKey("accounts.id"))

    company_name = Column(String(255))
    business_name = Column(String(255))
    account_status = Column(String(20), default="ACTIVE")

    # Billing
    billing_type = Column(String(20), default="PREPAID")
    credit_limit = Column(Numeric(15, 2), default=0)
    current_balance = Column(Numeric(15, 2), default=0)
    currency = Column(String(3), default="USD")
    low_balance_threshold = Column(Numeric(15, 2), default=100)

    # Sender configuration
    free_sender = Column(Boolean, default=False)
    allowed_sender_ids = Column(ARRAY(String))
    blocked_sender_ids = Column(ARRAY(String))
    default_sender_id = Column(String(20))

    # Routing
    bound_smsc_ids = Column(ARRAY(Integer))
    routing_policy = Column(String(50), default="ROUND_ROBIN")

    # Throughput limits
    max_tps = Column(Integer, default=100)
    max_concurrent_connections = Column(Integer, default=10)
    max_messages_per_day = Column(Integer)

    # Quotas
    daily_quota = Column(Integer)
    monthly_quota = Column(Integer)
    used_today = Column(Integer, default=0)
    used_this_month = Column(Integer, default=0)

    # Contact
    contact_name = Column(String(255))
    contact_email = Column(String(255))
    contact_phone = Column(String(50))
    address = Column(Text)

    # Timestamps
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow)
    expires_at = Column(DateTime)

    # Metadata
    metadata = Column(JSONB)

    # Relationships
    users = relationship("User", back_populates="account")
    parent = relationship("Account", remote_side=[id], backref="children")


class User(Base):
    """User model"""
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    user_id = Column(String(64), unique=True, nullable=False)
    account_id = Column(Integer, ForeignKey("accounts.id"))
    username = Column(String(100), unique=True, nullable=False)
    email = Column(String(255), unique=True, nullable=False)
    password_hash = Column(String(255), nullable=False)

    # Profile
    full_name = Column(String(255))
    phone = Column(String(50))
    language = Column(String(10), default="en")
    timezone = Column(String(50), default="UTC")

    # Status
    status = Column(String(20), default="ACTIVE")
    email_verified = Column(Boolean, default=False)
    phone_verified = Column(Boolean, default=False)
    must_change_password = Column(Boolean, default=True)

    # Security
    two_factor_enabled = Column(Boolean, default=False)
    two_factor_method = Column(String(20))
    two_factor_secret = Column(String(255))
    failed_login_attempts = Column(Integer, default=0)
    locked_until = Column(DateTime)
    password_changed_at = Column(DateTime)
    password_expires_at = Column(DateTime)

    # Session
    last_login_at = Column(DateTime)
    last_login_ip = Column(String(50))
    last_activity_at = Column(DateTime)

    # API Access
    api_enabled = Column(Boolean, default=False)
    api_key = Column(String(64), unique=True)
    api_key_expires_at = Column(DateTime)

    # SMPP Access
    smpp_enabled = Column(Boolean, default=False)
    smpp_username = Column(String(100))
    smpp_password = Column(String(255))
    smpp_system_type = Column(String(20))

    # Timestamps
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow)
    created_by = Column(Integer, ForeignKey("users.id"))

    # Relationships
    account = relationship("Account", back_populates="users")


class Role(Base):
    """Role model for RBAC"""
    __tablename__ = "roles"

    id = Column(Integer, primary_key=True)
    role_name = Column(String(50), unique=True, nullable=False)
    display_name = Column(String(100))
    description = Column(Text)
    is_system_role = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)


class Permission(Base):
    """Permission model for RBAC"""
    __tablename__ = "permissions"

    id = Column(Integer, primary_key=True)
    permission_key = Column(String(100), unique=True, nullable=False)
    module = Column(String(50))
    action = Column(String(50))
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
