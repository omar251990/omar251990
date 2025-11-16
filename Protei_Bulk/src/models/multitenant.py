#!/usr/bin/env python3
"""
Multi-Tenant Models
Customer, Role, Permission models for multi-tenant architecture
"""

from sqlalchemy import Column, Integer, BigInteger, String, Text, Boolean, TIMESTAMP, Numeric, ForeignKey, CheckConstraint, Index
from sqlalchemy.dialects.postgresql import JSONB, INET
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from datetime import datetime

from src.core.database import Base


class Customer(Base):
    """Customer (Tenant) Model"""
    __tablename__ = 'tbl_customers'

    customer_id = Column(BigInteger, primary_key=True, autoincrement=True)
    customer_code = Column(String(64), unique=True, nullable=False, index=True)
    customer_name = Column(String(255), nullable=False)
    company_name = Column(String(255))

    # Contact Information
    contact_person = Column(String(255))
    contact_email = Column(String(255))
    contact_phone = Column(String(50))

    # Address
    address = Column(Text)
    city = Column(String(100))
    country = Column(String(100))

    # Status & License
    status = Column(String(20), default='ACTIVE', nullable=False)
    license_key = Column(String(255), unique=True)
    license_type = Column(String(50), default='STANDARD')
    expiry_date = Column(TIMESTAMP)

    # Quotas & Limits
    max_users = Column(Integer, default=100)
    max_resellers = Column(Integer, default=10)
    max_tps = Column(Integer, default=1000)
    max_campaigns = Column(Integer, default=1000)
    storage_quota_gb = Column(Integer, default=100)

    # Allowed Features
    allowed_channels = Column(JSONB, default=["SMS", "EMAIL"])
    allowed_modules = Column(JSONB, default=["campaigns", "templates", "reports"])
    feature_flags = Column(JSONB, default={})

    # Branding
    logo_url = Column(String(500))
    primary_color = Column(String(20))
    secondary_color = Column(String(20))

    # Configuration
    timezone = Column(String(50), default='UTC')
    language = Column(String(10), default='EN')
    date_format = Column(String(20), default='YYYY-MM-DD')

    # SMSC & Routing
    allowed_smsc_ids = Column(JSONB, default=[])
    allowed_sender_ids = Column(JSONB, default=[])
    routing_policy = Column(String(50), default='AUTOMATIC')

    # Billing
    billing_type = Column(String(20), default='PREPAID')
    balance = Column(Numeric(15, 4), default=0.0)
    credit_limit = Column(Numeric(15, 4), default=0.0)
    cost_per_sms = Column(Numeric(10, 4), default=0.01)

    # Alert Thresholds
    alert_balance_threshold = Column(Numeric(15, 4), default=1000.0)
    alert_tps_threshold = Column(Integer, default=900)
    alert_emails = Column(JSONB, default=[])

    # Audit
    created_by = Column(String(64))
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())
    last_activity_at = Column(TIMESTAMP)

    # Metadata
    metadata = Column(JSONB, default={})
    notes = Column(Text)

    # Relationships
    users = relationship('User', back_populates='customer')
    campaigns = relationship('Campaign', back_populates='customer')
    configs = relationship('CustomerConfig', back_populates='customer')

    __table_args__ = (
        CheckConstraint("status IN ('ACTIVE', 'SUSPENDED', 'EXPIRED', 'TRIAL')", name='check_customer_status'),
        CheckConstraint("license_type IN ('TRIAL', 'STANDARD', 'PREMIUM', 'ENTERPRISE')", name='check_license_type'),
        CheckConstraint("billing_type IN ('PREPAID', 'POSTPAID')", name='check_billing_type'),
        Index('idx_customers_status', 'status'),
        Index('idx_customers_expiry', 'expiry_date'),
    )

    def __repr__(self):
        return f"<Customer(id={self.customer_id}, name='{self.customer_name}', status='{self.status}')>"


class Role(Base):
    """Role Model"""
    __tablename__ = 'tbl_roles'

    role_id = Column(BigInteger, primary_key=True, autoincrement=True)
    role_name = Column(String(100), nullable=False)
    role_code = Column(String(50), unique=True, nullable=False)
    description = Column(Text)

    # Hierarchy Level
    level = Column(Integer, default=100)  # Lower = higher privilege

    # Scope
    scope = Column(String(20), default='CUSTOMER')

    # Status
    is_system_role = Column(Boolean, default=False)
    is_active = Column(Boolean, default=True)

    # Audit
    created_at = Column(TIMESTAMP, default=func.now())
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Relationships
    users = relationship('User', back_populates='role')
    permissions = relationship('RolePermission', back_populates='role')

    __table_args__ = (
        CheckConstraint("scope IN ('GLOBAL', 'CUSTOMER', 'RESELLER')", name='check_role_scope'),
    )

    def __repr__(self):
        return f"<Role(id={self.role_id}, name='{self.role_name}', level={self.level})>"


class Permission(Base):
    """Permission Model"""
    __tablename__ = 'tbl_permissions'

    permission_id = Column(BigInteger, primary_key=True, autoincrement=True)
    permission_code = Column(String(100), unique=True, nullable=False)
    module_name = Column(String(100), nullable=False)
    action_name = Column(String(100), nullable=False)
    description = Column(Text)

    # Categorization
    category = Column(String(50))

    # Status
    is_system_permission = Column(Boolean, default=True)
    is_active = Column(Boolean, default=True)

    # Audit
    created_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    role_permissions = relationship('RolePermission', back_populates='permission')
    user_overrides = relationship('UserPermissionOverride', back_populates='permission')

    def __repr__(self):
        return f"<Permission(id={self.permission_id}, code='{self.permission_code}')>"


class RolePermission(Base):
    """Role Permission Mapping"""
    __tablename__ = 'tbl_role_permissions'

    role_id = Column(BigInteger, ForeignKey('tbl_roles.role_id', ondelete='CASCADE'), primary_key=True)
    permission_id = Column(BigInteger, ForeignKey('tbl_permissions.permission_id', ondelete='CASCADE'), primary_key=True)
    allow = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, default=func.now())

    # Relationships
    role = relationship('Role', back_populates='permissions')
    permission = relationship('Permission', back_populates='role_permissions')

    __table_args__ = (
        Index('idx_role_permissions_role', 'role_id'),
        Index('idx_role_permissions_permission', 'permission_id'),
    )

    def __repr__(self):
        return f"<RolePermission(role_id={self.role_id}, permission_id={self.permission_id}, allow={self.allow})>"


class UserPermissionOverride(Base):
    """User Permission Override"""
    __tablename__ = 'tbl_user_permissions_override'

    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='CASCADE'), primary_key=True)
    permission_id = Column(BigInteger, ForeignKey('tbl_permissions.permission_id', ondelete='CASCADE'), primary_key=True)
    allow = Column(Boolean, default=True)
    granted_by = Column(String(64))
    granted_at = Column(TIMESTAMP, default=func.now())
    expires_at = Column(TIMESTAMP)
    reason = Column(Text)

    # Relationships
    user = relationship('User', back_populates='permission_overrides')
    permission = relationship('Permission', back_populates='user_overrides')

    __table_args__ = (
        Index('idx_user_permissions_user', 'user_id'),
        Index('idx_user_permissions_permission', 'permission_id'),
        Index('idx_user_permissions_expires', 'expires_at'),
    )

    def __repr__(self):
        return f"<UserPermissionOverride(user_id={self.user_id}, permission_id={self.permission_id}, allow={self.allow})>"


class CustomerConfig(Base):
    """Customer Configuration"""
    __tablename__ = 'tbl_customer_config'

    config_id = Column(BigInteger, primary_key=True, autoincrement=True)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'), nullable=False)
    config_key = Column(String(100), nullable=False)
    config_value = Column(Text)
    config_type = Column(String(50), default='STRING')
    is_encrypted = Column(Boolean, default=False)
    description = Column(Text)
    updated_at = Column(TIMESTAMP, default=func.now(), onupdate=func.now())

    # Relationships
    customer = relationship('Customer', back_populates='configs')

    __table_args__ = (
        Index('idx_customer_config_customer', 'customer_id'),
        Index('idx_customer_config_key', 'config_key'),
    )

    def __repr__(self):
        return f"<CustomerConfig(customer_id={self.customer_id}, key='{self.config_key}')>"


class PermissionAudit(Base):
    """Permission Audit Log"""
    __tablename__ = 'tbl_permission_audit'

    audit_id = Column(BigInteger, primary_key=True, autoincrement=True)
    customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))
    target_user_id = Column(BigInteger, ForeignKey('users.id', ondelete='SET NULL'))
    permission_id = Column(BigInteger, ForeignKey('tbl_permissions.permission_id', ondelete='SET NULL'))
    action = Column(String(50), nullable=False)  # GRANTED, REVOKED, CHECKED
    result = Column(Boolean)
    ip_address = Column(INET)
    user_agent = Column(Text)
    created_at = Column(TIMESTAMP, default=func.now())
    metadata = Column(JSONB, default={})

    __table_args__ = (
        Index('idx_permission_audit_customer', 'customer_id'),
        Index('idx_permission_audit_user', 'user_id'),
        Index('idx_permission_audit_created', 'created_at'),
    )

    def __repr__(self):
        return f"<PermissionAudit(id={self.audit_id}, action='{self.action}', result={self.result})>"


# Update User model to include multi-tenant fields
def extend_user_model():
    """
    Extend existing User model with multi-tenant fields
    This should be called after the User model is defined
    """
    from src.models.user import User

    # Add new columns
    User.customer_id = Column(BigInteger, ForeignKey('tbl_customers.customer_id', ondelete='CASCADE'))
    User.role_id = Column(BigInteger, ForeignKey('tbl_roles.role_id'))
    User.parent_user_id = Column(BigInteger, ForeignKey('users.id'))
    User.user_level = Column(Integer, default=30)

    # Add relationships
    User.customer = relationship('Customer', back_populates='users')
    User.role = relationship('Role', back_populates='users')
    User.parent_user = relationship('User', remote_side=[User.id], backref='sub_users')
    User.permission_overrides = relationship('UserPermissionOverride', back_populates='user')

    # Add indexes
    Index('idx_users_customer', User.customer_id)
    Index('idx_users_role', User.role_id)
    Index('idx_users_parent', User.parent_user_id)
