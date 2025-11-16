#!/usr/bin/env python3
"""
Customer Service
Handles customer (tenant) management operations
"""

from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_, func

from src.models.user import User
from src.models.multitenant import Customer, Role, CustomerConfig


class CustomerService:
    """
    Customer Service for managing tenants
    """

    def __init__(self):
        pass

    def create_customer(
        self,
        db: Session,
        customer_code: str,
        customer_name: str,
        company_name: str = None,
        contact_email: str = None,
        license_type: str = 'STANDARD',
        max_users: int = 100,
        max_tps: int = 1000,
        billing_type: str = 'PREPAID',
        created_by: str = None,
        **kwargs
    ) -> Customer:
        """
        Create a new customer (tenant)

        Args:
            db: Database session
            customer_code: Unique customer code
            customer_name: Customer display name
            company_name: Company name
            contact_email: Contact email
            license_type: License type (TRIAL, STANDARD, PREMIUM, ENTERPRISE)
            max_users: Maximum users allowed
            max_tps: Maximum TPS limit
            billing_type: PREPAID or POSTPAID
            created_by: Username of creator
            **kwargs: Additional customer attributes

        Returns:
            Customer: Created customer object
        """
        # Check if customer code already exists
        existing = db.query(Customer).filter(Customer.customer_code == customer_code).first()
        if existing:
            raise ValueError(f"Customer with code '{customer_code}' already exists")

        # Set expiry date based on license type
        if license_type == 'TRIAL':
            expiry_date = datetime.utcnow() + timedelta(days=30)
        else:
            expiry_date = datetime.utcnow() + timedelta(days=365)

        # Create customer
        customer = Customer(
            customer_code=customer_code,
            customer_name=customer_name,
            company_name=company_name or customer_name,
            contact_email=contact_email,
            license_type=license_type,
            max_users=max_users,
            max_tps=max_tps,
            billing_type=billing_type,
            expiry_date=expiry_date,
            status='ACTIVE',
            created_by=created_by,
            **kwargs
        )

        db.add(customer)
        db.commit()
        db.refresh(customer)

        return customer

    def update_customer(
        self,
        db: Session,
        customer_id: int,
        **kwargs
    ) -> Optional[Customer]:
        """
        Update customer details

        Args:
            db: Database session
            customer_id: Customer ID
            **kwargs: Fields to update

        Returns:
            Customer: Updated customer object
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return None

        # Update fields
        for key, value in kwargs.items():
            if hasattr(customer, key):
                setattr(customer, key, value)

        customer.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(customer)

        return customer

    def get_customer(
        self,
        db: Session,
        customer_id: int = None,
        customer_code: str = None
    ) -> Optional[Customer]:
        """
        Get customer by ID or code

        Args:
            db: Database session
            customer_id: Customer ID
            customer_code: Customer code

        Returns:
            Customer: Customer object
        """
        if customer_id:
            return db.query(Customer).filter(Customer.customer_id == customer_id).first()
        elif customer_code:
            return db.query(Customer).filter(Customer.customer_code == customer_code).first()
        return None

    def list_customers(
        self,
        db: Session,
        status: str = None,
        license_type: str = None,
        limit: int = 100,
        offset: int = 0
    ) -> List[Customer]:
        """
        List all customers with optional filtering

        Args:
            db: Database session
            status: Filter by status
            license_type: Filter by license type
            limit: Maximum records to return
            offset: Number of records to skip

        Returns:
            List[Customer]: List of customers
        """
        query = db.query(Customer)

        if status:
            query = query.filter(Customer.status == status)

        if license_type:
            query = query.filter(Customer.license_type == license_type)

        return query.order_by(Customer.created_at.desc()).limit(limit).offset(offset).all()

    def suspend_customer(
        self,
        db: Session,
        customer_id: int,
        reason: str = None
    ) -> bool:
        """
        Suspend a customer (blocks all access)

        Args:
            db: Database session
            customer_id: Customer ID
            reason: Reason for suspension

        Returns:
            bool: True if successful
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return False

        customer.status = 'SUSPENDED'
        customer.updated_at = datetime.utcnow()

        if reason:
            customer.notes = f"{customer.notes}\n[{datetime.utcnow()}] SUSPENDED: {reason}" if customer.notes else f"SUSPENDED: {reason}"

        # Deactivate all users under this customer
        db.query(User).filter(User.customer_id == customer_id).update({'is_active': False})

        db.commit()
        return True

    def activate_customer(
        self,
        db: Session,
        customer_id: int
    ) -> bool:
        """
        Activate a suspended customer

        Args:
            db: Database session
            customer_id: Customer ID

        Returns:
            bool: True if successful
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return False

        customer.status = 'ACTIVE'
        customer.updated_at = datetime.utcnow()

        # Reactivate customer admin
        db.query(User).filter(
            and_(
                User.customer_id == customer_id,
                User.role_id == db.query(Role.role_id).filter(Role.role_code == 'CUSTOMER_ADMIN').scalar()
            )
        ).update({'is_active': True})

        db.commit()
        return True

    def delete_customer(
        self,
        db: Session,
        customer_id: int
    ) -> bool:
        """
        Delete a customer (permanent, use with caution!)

        Args:
            db: Database session
            customer_id: Customer ID

        Returns:
            bool: True if successful
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return False

        # Due to CASCADE, this will delete all related data
        db.delete(customer)
        db.commit()
        return True

    def get_customer_statistics(
        self,
        db: Session,
        customer_id: int
    ) -> Dict[str, Any]:
        """
        Get customer usage statistics

        Args:
            db: Database session
            customer_id: Customer ID

        Returns:
            Dict: Statistics dictionary
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return {}

        # Count users
        total_users = db.query(func.count(User.id)).filter(User.customer_id == customer_id).scalar()
        active_users = db.query(func.count(User.id)).filter(
            and_(User.customer_id == customer_id, User.is_active == True)
        ).scalar()

        # Count campaigns
        from src.models.campaign import Campaign
        total_campaigns = db.query(func.count(Campaign.id)).filter(Campaign.customer_id == customer_id).scalar() or 0

        # Count messages
        from src.models.message import Message
        total_messages = db.query(func.count(Message.id)).filter(Message.customer_id == customer_id).scalar() or 0

        # Get today's message count
        today_start = datetime.utcnow().replace(hour=0, minute=0, second=0, microsecond=0)
        messages_today = db.query(func.count(Message.id)).filter(
            and_(
                Message.customer_id == customer_id,
                Message.created_at >= today_start
            )
        ).scalar() or 0

        return {
            'customer_id': customer_id,
            'customer_name': customer.customer_name,
            'status': customer.status,
            'license_type': customer.license_type,
            'expiry_date': customer.expiry_date.isoformat() if customer.expiry_date else None,
            'users': {
                'total': total_users,
                'active': active_users,
                'max_allowed': customer.max_users,
                'utilization_percent': (total_users / customer.max_users * 100) if customer.max_users > 0 else 0
            },
            'campaigns': {
                'total': total_campaigns,
                'max_allowed': customer.max_campaigns
            },
            'messages': {
                'total': total_messages,
                'today': messages_today
            },
            'quota': {
                'max_tps': customer.max_tps,
                'storage_gb': customer.storage_quota_gb
            },
            'billing': {
                'type': customer.billing_type,
                'balance': float(customer.balance),
                'credit_limit': float(customer.credit_limit)
            },
            'last_activity': customer.last_activity_at.isoformat() if customer.last_activity_at else None
        }

    def set_customer_config(
        self,
        db: Session,
        customer_id: int,
        config_key: str,
        config_value: str,
        config_type: str = 'STRING',
        is_encrypted: bool = False,
        description: str = None
    ) -> CustomerConfig:
        """
        Set a customer configuration value

        Args:
            db: Database session
            customer_id: Customer ID
            config_key: Configuration key
            config_value: Configuration value
            config_type: Value type (STRING, INTEGER, BOOLEAN, JSON)
            is_encrypted: Whether value should be encrypted
            description: Description of config

        Returns:
            CustomerConfig: Configuration object
        """
        # Check if config exists
        config = db.query(CustomerConfig).filter(
            and_(
                CustomerConfig.customer_id == customer_id,
                CustomerConfig.config_key == config_key
            )
        ).first()

        if config:
            config.config_value = config_value
            config.config_type = config_type
            config.is_encrypted = is_encrypted
            config.description = description
            config.updated_at = datetime.utcnow()
        else:
            config = CustomerConfig(
                customer_id=customer_id,
                config_key=config_key,
                config_value=config_value,
                config_type=config_type,
                is_encrypted=is_encrypted,
                description=description
            )
            db.add(config)

        db.commit()
        db.refresh(config)
        return config

    def get_customer_config(
        self,
        db: Session,
        customer_id: int,
        config_key: str
    ) -> Optional[CustomerConfig]:
        """
        Get a customer configuration value

        Args:
            db: Database session
            customer_id: Customer ID
            config_key: Configuration key

        Returns:
            CustomerConfig: Configuration object
        """
        return db.query(CustomerConfig).filter(
            and_(
                CustomerConfig.customer_id == customer_id,
                CustomerConfig.config_key == config_key
            )
        ).first()

    def get_customer_configs(
        self,
        db: Session,
        customer_id: int
    ) -> List[CustomerConfig]:
        """
        Get all configurations for a customer

        Args:
            db: Database session
            customer_id: Customer ID

        Returns:
            List[CustomerConfig]: List of configurations
        """
        return db.query(CustomerConfig).filter(
            CustomerConfig.customer_id == customer_id
        ).all()

    def update_customer_balance(
        self,
        db: Session,
        customer_id: int,
        amount: float,
        operation: str = 'ADD'
    ) -> Optional[Customer]:
        """
        Update customer balance

        Args:
            db: Database session
            customer_id: Customer ID
            amount: Amount to add or subtract
            operation: 'ADD' or 'SUBTRACT'

        Returns:
            Customer: Updated customer
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return None

        if operation == 'ADD':
            customer.balance += amount
        elif operation == 'SUBTRACT':
            customer.balance -= amount

        customer.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(customer)

        return customer

    def extend_license(
        self,
        db: Session,
        customer_id: int,
        days: int
    ) -> Optional[Customer]:
        """
        Extend customer license expiry date

        Args:
            db: Database session
            customer_id: Customer ID
            days: Number of days to extend

        Returns:
            Customer: Updated customer
        """
        customer = db.query(Customer).filter(Customer.customer_id == customer_id).first()
        if not customer:
            return None

        if customer.expiry_date:
            customer.expiry_date += timedelta(days=days)
        else:
            customer.expiry_date = datetime.utcnow() + timedelta(days=days)

        customer.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(customer)

        return customer


# Global customer service instance
customer_service = CustomerService()
