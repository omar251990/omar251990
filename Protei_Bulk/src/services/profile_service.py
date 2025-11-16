#!/usr/bin/env python3
"""
Profile Management Service
Handles subscriber profiles with privacy-first design
"""

import hashlib
import logging
from typing import List, Optional, Dict, Any, Tuple
from datetime import datetime, date
from sqlalchemy import and_, or_, func, cast, String
from sqlalchemy.orm import Session
from sqlalchemy.dialects.postgresql import JSONB

from src.models.profiling import (
    Profile, ProfileGroup, ProfileGroupMember,
    AttributeSchema, ProfileImportJob, ProfileQueryLog, ProfileStatistics
)
from src.models.customer import Customer
from src.models.user import User

logger = logging.getLogger(__name__)


class ProfileService:
    """Service for managing subscriber profiles"""

    @staticmethod
    def hash_msisdn(msisdn: str) -> str:
        """
        Hash MSISDN using SHA256 for privacy protection

        Args:
            msisdn: Plain text MSISDN

        Returns:
            SHA256 hash of the MSISDN
        """
        # Normalize MSISDN (remove spaces, dashes, etc.)
        normalized = ''.join(filter(str.isdigit, msisdn))

        # Hash using SHA256
        return hashlib.sha256(normalized.encode()).hexdigest()

    def create_profile(
        self,
        db: Session,
        msisdn: str,
        customer_id: int,
        profile_data: Dict[str, Any],
        user_id: Optional[int] = None
    ) -> Profile:
        """
        Create a new subscriber profile

        Args:
            db: Database session
            msisdn: Subscriber MSISDN (will be hashed)
            customer_id: Customer ID
            profile_data: Profile attributes
            user_id: User creating the profile

        Returns:
            Created Profile object
        """
        try:
            # Hash MSISDN for privacy
            msisdn_hash = self.hash_msisdn(msisdn)

            # Check if profile already exists
            existing = db.query(Profile).filter(
                Profile.msisdn_hash == msisdn_hash
            ).first()

            if existing:
                logger.warning(f"Profile already exists for MSISDN hash: {msisdn_hash}")
                return existing

            # Separate standard and custom attributes
            standard_fields = {
                'gender', 'age', 'date_of_birth', 'language',
                'country_code', 'region', 'city', 'postal_code',
                'device_type', 'device_model', 'os_version',
                'plan_type', 'subscription_date', 'last_recharge_date',
                'interests', 'preferences', 'status',
                'opt_in_marketing', 'opt_in_sms'
            }

            standard_attrs = {k: v for k, v in profile_data.items() if k in standard_fields}
            custom_attrs = {k: v for k, v in profile_data.items() if k not in standard_fields}

            # Create profile
            profile = Profile(
                msisdn_hash=msisdn_hash,
                customer_id=customer_id,
                custom_attributes=custom_attrs,
                imported_by=str(user_id) if user_id else None,
                imported_at=datetime.utcnow(),
                **standard_attrs
            )

            db.add(profile)
            db.commit()
            db.refresh(profile)

            logger.info(f"Created profile {profile.profile_id} for customer {customer_id}")
            return profile

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating profile: {str(e)}")
            raise

    def get_profile(
        self,
        db: Session,
        profile_id: Optional[int] = None,
        msisdn: Optional[str] = None
    ) -> Optional[Profile]:
        """
        Get profile by ID or MSISDN

        Args:
            db: Database session
            profile_id: Profile ID
            msisdn: MSISDN (will be hashed for lookup)

        Returns:
            Profile object or None
        """
        try:
            if profile_id:
                return db.query(Profile).filter(Profile.profile_id == profile_id).first()
            elif msisdn:
                msisdn_hash = self.hash_msisdn(msisdn)
                return db.query(Profile).filter(Profile.msisdn_hash == msisdn_hash).first()
            return None

        except Exception as e:
            logger.error(f"Error fetching profile: {str(e)}")
            raise

    def update_profile(
        self,
        db: Session,
        profile_id: int,
        profile_data: Dict[str, Any],
        user_id: Optional[int] = None
    ) -> Optional[Profile]:
        """
        Update existing profile

        Args:
            db: Database session
            profile_id: Profile ID to update
            profile_data: Updated attributes
            user_id: User making the update

        Returns:
            Updated Profile object or None
        """
        try:
            profile = db.query(Profile).filter(Profile.profile_id == profile_id).first()

            if not profile:
                logger.warning(f"Profile {profile_id} not found")
                return None

            # Separate standard and custom attributes
            standard_fields = {
                'gender', 'age', 'date_of_birth', 'language',
                'country_code', 'region', 'city', 'postal_code',
                'device_type', 'device_model', 'os_version',
                'plan_type', 'subscription_date', 'last_recharge_date',
                'interests', 'preferences', 'status',
                'opt_in_marketing', 'opt_in_sms'
            }

            # Update standard fields
            for field, value in profile_data.items():
                if field in standard_fields and hasattr(profile, field):
                    setattr(profile, field, value)
                else:
                    # Add to custom attributes
                    if profile.custom_attributes is None:
                        profile.custom_attributes = {}
                    profile.custom_attributes[field] = value

            profile.updated_at = datetime.utcnow()

            db.commit()
            db.refresh(profile)

            logger.info(f"Updated profile {profile_id}")
            return profile

        except Exception as e:
            db.rollback()
            logger.error(f"Error updating profile: {str(e)}")
            raise

    def delete_profile(
        self,
        db: Session,
        profile_id: int,
        soft_delete: bool = True
    ) -> bool:
        """
        Delete profile (soft or hard delete)

        Args:
            db: Database session
            profile_id: Profile ID to delete
            soft_delete: If True, mark as deleted; if False, remove from database

        Returns:
            True if successful
        """
        try:
            profile = db.query(Profile).filter(Profile.profile_id == profile_id).first()

            if not profile:
                logger.warning(f"Profile {profile_id} not found")
                return False

            if soft_delete:
                # Soft delete: mark as deleted
                profile.status = 'DELETED'
                profile.updated_at = datetime.utcnow()
                db.commit()
                logger.info(f"Soft deleted profile {profile_id}")
            else:
                # Hard delete: remove from database
                db.delete(profile)
                db.commit()
                logger.info(f"Hard deleted profile {profile_id}")

            return True

        except Exception as e:
            db.rollback()
            logger.error(f"Error deleting profile: {str(e)}")
            raise

    def search_profiles(
        self,
        db: Session,
        customer_id: int,
        filters: Dict[str, Any],
        user_id: Optional[int] = None,
        offset: int = 0,
        limit: int = 100
    ) -> Tuple[List[Profile], int]:
        """
        Search profiles with filters

        Args:
            db: Database session
            customer_id: Customer ID
            filters: Search filters
            user_id: User performing search
            offset: Pagination offset
            limit: Results per page

        Returns:
            Tuple of (profiles list, total count)
        """
        try:
            # Start with base query
            query = db.query(Profile).filter(Profile.customer_id == customer_id)

            # Apply filters
            if 'gender' in filters:
                query = query.filter(Profile.gender == filters['gender'])

            if 'age_min' in filters:
                query = query.filter(Profile.age >= filters['age_min'])

            if 'age_max' in filters:
                query = query.filter(Profile.age <= filters['age_max'])

            if 'region' in filters:
                query = query.filter(Profile.region == filters['region'])

            if 'city' in filters:
                query = query.filter(Profile.city == filters['city'])

            if 'device_type' in filters:
                query = query.filter(Profile.device_type == filters['device_type'])

            if 'plan_type' in filters:
                query = query.filter(Profile.plan_type == filters['plan_type'])

            if 'status' in filters:
                query = query.filter(Profile.status == filters['status'])

            if 'opt_in_marketing' in filters:
                query = query.filter(Profile.opt_in_marketing == filters['opt_in_marketing'])

            if 'last_activity_after' in filters:
                query = query.filter(Profile.last_activity_date >= filters['last_activity_after'])

            if 'last_activity_before' in filters:
                query = query.filter(Profile.last_activity_date <= filters['last_activity_before'])

            # Custom attributes search (JSONB)
            if 'custom_attributes' in filters:
                for key, value in filters['custom_attributes'].items():
                    query = query.filter(
                        Profile.custom_attributes[key].astext == str(value)
                    )

            # Get total count
            total = query.count()

            # Apply pagination
            profiles = query.offset(offset).limit(limit).all()

            # Log query for privacy audit
            self._log_query(
                db=db,
                customer_id=customer_id,
                user_id=user_id,
                query_type='SEARCH',
                filter_query=filters,
                result_count=total
            )

            logger.info(f"Search returned {len(profiles)} profiles (total: {total})")
            return profiles, total

        except Exception as e:
            logger.error(f"Error searching profiles: {str(e)}")
            raise

    def bulk_update_profiles(
        self,
        db: Session,
        profile_ids: List[int],
        update_data: Dict[str, Any],
        user_id: Optional[int] = None
    ) -> int:
        """
        Bulk update multiple profiles

        Args:
            db: Database session
            profile_ids: List of profile IDs to update
            update_data: Data to update
            user_id: User performing update

        Returns:
            Number of profiles updated
        """
        try:
            count = 0

            for profile_id in profile_ids:
                profile = self.update_profile(db, profile_id, update_data, user_id)
                if profile:
                    count += 1

            logger.info(f"Bulk updated {count} profiles")
            return count

        except Exception as e:
            logger.error(f"Error in bulk update: {str(e)}")
            raise

    def get_profile_statistics(
        self,
        db: Session,
        customer_id: int,
        period_date: Optional[date] = None
    ) -> Optional[ProfileStatistics]:
        """
        Get profile statistics for a customer

        Args:
            db: Database session
            customer_id: Customer ID
            period_date: Statistics date (defaults to today)

        Returns:
            ProfileStatistics object or None
        """
        try:
            if not period_date:
                period_date = date.today()

            stats = db.query(ProfileStatistics).filter(
                and_(
                    ProfileStatistics.customer_id == customer_id,
                    ProfileStatistics.period_date == period_date,
                    ProfileStatistics.period_type == 'DAILY'
                )
            ).first()

            return stats

        except Exception as e:
            logger.error(f"Error fetching statistics: {str(e)}")
            raise

    def calculate_profile_statistics(
        self,
        db: Session,
        customer_id: int,
        period_date: Optional[date] = None
    ) -> ProfileStatistics:
        """
        Calculate and store profile statistics

        Args:
            db: Database session
            customer_id: Customer ID
            period_date: Statistics date (defaults to today)

        Returns:
            ProfileStatistics object
        """
        try:
            if not period_date:
                period_date = date.today()

            # Get existing stats or create new
            stats = db.query(ProfileStatistics).filter(
                and_(
                    ProfileStatistics.customer_id == customer_id,
                    ProfileStatistics.period_date == period_date,
                    ProfileStatistics.period_type == 'DAILY'
                )
            ).first()

            if not stats:
                stats = ProfileStatistics(
                    customer_id=customer_id,
                    period_date=period_date,
                    period_type='DAILY'
                )
                db.add(stats)

            # Calculate counts
            base_query = db.query(Profile).filter(Profile.customer_id == customer_id)

            stats.total_profiles = base_query.count()
            stats.active_profiles = base_query.filter(Profile.status == 'ACTIVE').count()
            stats.inactive_profiles = base_query.filter(Profile.status == 'INACTIVE').count()

            # Demographics
            stats.male_count = base_query.filter(Profile.gender == 'MALE').count()
            stats.female_count = base_query.filter(Profile.gender == 'FEMALE').count()

            # Average age
            avg_age = db.query(func.avg(Profile.age)).filter(
                and_(
                    Profile.customer_id == customer_id,
                    Profile.age.isnot(None)
                )
            ).scalar()
            stats.avg_age = float(avg_age) if avg_age else None

            # Device distribution
            stats.android_count = base_query.filter(Profile.device_type == 'ANDROID').count()
            stats.ios_count = base_query.filter(Profile.device_type == 'IOS').count()
            stats.feature_phone_count = base_query.filter(Profile.device_type == 'FEATURE_PHONE').count()

            # Plan distribution
            stats.prepaid_count = base_query.filter(Profile.plan_type == 'PREPAID').count()
            stats.postpaid_count = base_query.filter(Profile.plan_type == 'POSTPAID').count()

            # Opt-in rates
            stats.opt_in_marketing_count = base_query.filter(Profile.opt_in_marketing == True).count()
            stats.opt_in_sms_count = base_query.filter(Profile.opt_in_sms == True).count()

            db.commit()
            db.refresh(stats)

            logger.info(f"Calculated statistics for customer {customer_id}: {stats.total_profiles} profiles")
            return stats

        except Exception as e:
            db.rollback()
            logger.error(f"Error calculating statistics: {str(e)}")
            raise

    def _log_query(
        self,
        db: Session,
        customer_id: int,
        user_id: Optional[int],
        query_type: str,
        filter_query: Dict[str, Any],
        result_count: int,
        group_id: Optional[int] = None,
        group_name: Optional[str] = None
    ) -> None:
        """
        Log query for privacy audit trail

        Args:
            db: Database session
            customer_id: Customer ID
            user_id: User ID
            query_type: Type of query
            filter_query: Query filters
            result_count: Number of results
            group_id: Profile group ID (if applicable)
            group_name: Profile group name (if applicable)
        """
        try:
            log = ProfileQueryLog(
                customer_id=customer_id,
                user_id=user_id,
                query_type=query_type,
                filter_query=filter_query,
                result_count=result_count,
                group_id=group_id,
                group_name=group_name,
                includes_pii=False  # We use hashed MSISDNs
            )

            db.add(log)
            db.commit()

        except Exception as e:
            logger.error(f"Error logging query: {str(e)}")
            # Don't raise - logging failure shouldn't break the operation


class AttributeSchemaService:
    """Service for managing dynamic attribute schemas"""

    def create_attribute(
        self,
        db: Session,
        attribute_data: Dict[str, Any],
        user_id: Optional[str] = None
    ) -> AttributeSchema:
        """
        Create a new attribute schema

        Args:
            db: Database session
            attribute_data: Attribute configuration
            user_id: User creating the attribute

        Returns:
            Created AttributeSchema object
        """
        try:
            attribute = AttributeSchema(
                created_by=user_id,
                **attribute_data
            )

            db.add(attribute)
            db.commit()
            db.refresh(attribute)

            logger.info(f"Created attribute schema: {attribute.attribute_name}")
            return attribute

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating attribute: {str(e)}")
            raise

    def get_attributes(
        self,
        db: Session,
        active_only: bool = True
    ) -> List[AttributeSchema]:
        """
        Get all attribute schemas

        Args:
            db: Database session
            active_only: If True, return only active attributes

        Returns:
            List of AttributeSchema objects
        """
        try:
            query = db.query(AttributeSchema)

            if active_only:
                query = query.filter(AttributeSchema.is_active == True)

            attributes = query.order_by(AttributeSchema.display_order).all()

            logger.info(f"Retrieved {len(attributes)} attributes")
            return attributes

        except Exception as e:
            logger.error(f"Error fetching attributes: {str(e)}")
            raise

    def validate_attribute_value(
        self,
        db: Session,
        attribute_code: str,
        value: Any
    ) -> Tuple[bool, Optional[str]]:
        """
        Validate attribute value against schema

        Args:
            db: Database session
            attribute_code: Attribute code
            value: Value to validate

        Returns:
            Tuple of (is_valid, error_message)
        """
        try:
            attribute = db.query(AttributeSchema).filter(
                AttributeSchema.attribute_code == attribute_code
            ).first()

            if not attribute:
                return False, f"Unknown attribute: {attribute_code}"

            # Type validation
            if attribute.data_type == 'STRING' and not isinstance(value, str):
                return False, f"{attribute_code} must be a string"

            elif attribute.data_type == 'INTEGER' and not isinstance(value, int):
                return False, f"{attribute_code} must be an integer"

            elif attribute.data_type == 'DECIMAL' and not isinstance(value, (int, float)):
                return False, f"{attribute_code} must be a number"

            elif attribute.data_type == 'BOOLEAN' and not isinstance(value, bool):
                return False, f"{attribute_code} must be a boolean"

            # Enum validation
            if attribute.data_type == 'ENUM' and attribute.allowed_values:
                if value not in attribute.allowed_values:
                    return False, f"{attribute_code} must be one of: {', '.join(map(str, attribute.allowed_values))}"

            # Range validation
            if attribute.min_value is not None and value < attribute.min_value:
                return False, f"{attribute_code} must be >= {attribute.min_value}"

            if attribute.max_value is not None and value > attribute.max_value:
                return False, f"{attribute_code} must be <= {attribute.max_value}"

            return True, None

        except Exception as e:
            logger.error(f"Error validating attribute: {str(e)}")
            return False, str(e)
