#!/usr/bin/env python3
"""
Segmentation Engine Service
Dynamic audience segmentation with query builder and privacy layer
"""

import logging
from typing import List, Optional, Dict, Any, Tuple
from datetime import datetime, timedelta
from sqlalchemy import and_, or_, not_, func, cast, String, text
from sqlalchemy.orm import Session
from sqlalchemy.dialects.postgresql import JSONB

from src.models.profiling import (
    Profile, ProfileGroup, ProfileGroupMember, ProfileQueryLog
)
from src.services.profile_service import ProfileService

logger = logging.getLogger(__name__)


class SegmentationService:
    """Service for managing profile segments and audiences"""

    def __init__(self):
        self.profile_service = ProfileService()

    def create_segment(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        group_data: Dict[str, Any]
    ) -> ProfileGroup:
        """
        Create a new profile segment

        Args:
            db: Database session
            customer_id: Customer ID
            user_id: User creating the segment
            group_data: Segment configuration

        Returns:
            Created ProfileGroup object
        """
        try:
            # Generate unique code if not provided
            if 'group_code' not in group_data:
                group_data['group_code'] = self._generate_group_code(db)

            # Create segment
            segment = ProfileGroup(
                customer_id=customer_id,
                user_id=user_id,
                created_by=str(user_id),
                **group_data
            )

            # Generate SQL from filter query
            if segment.filter_query:
                segment.filter_sql = self._build_sql_from_filters(segment.filter_query)

            db.add(segment)
            db.commit()
            db.refresh(segment)

            # Calculate initial membership if dynamic
            if segment.is_dynamic:
                self.refresh_segment(db, segment.group_id)

            logger.info(f"Created segment {segment.group_id}: {segment.group_name}")
            return segment

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating segment: {str(e)}")
            raise

    def get_segment(
        self,
        db: Session,
        group_id: Optional[int] = None,
        group_code: Optional[str] = None
    ) -> Optional[ProfileGroup]:
        """
        Get segment by ID or code

        Args:
            db: Database session
            group_id: Segment ID
            group_code: Segment code

        Returns:
            ProfileGroup object or None
        """
        try:
            if group_id:
                return db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()
            elif group_code:
                return db.query(ProfileGroup).filter(ProfileGroup.group_code == group_code).first()
            return None

        except Exception as e:
            logger.error(f"Error fetching segment: {str(e)}")
            raise

    def update_segment(
        self,
        db: Session,
        group_id: int,
        group_data: Dict[str, Any],
        user_id: Optional[int] = None
    ) -> Optional[ProfileGroup]:
        """
        Update existing segment

        Args:
            db: Database session
            group_id: Segment ID
            group_data: Updated configuration
            user_id: User making update

        Returns:
            Updated ProfileGroup object or None
        """
        try:
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                logger.warning(f"Segment {group_id} not found")
                return None

            # Update fields
            for field, value in group_data.items():
                if hasattr(segment, field):
                    setattr(segment, field, value)

            # Regenerate SQL if filter changed
            if 'filter_query' in group_data:
                segment.filter_sql = self._build_sql_from_filters(segment.filter_query)

            segment.updated_at = datetime.utcnow()

            db.commit()
            db.refresh(segment)

            # Refresh membership if dynamic
            if segment.is_dynamic:
                self.refresh_segment(db, segment.group_id)

            logger.info(f"Updated segment {group_id}")
            return segment

        except Exception as e:
            db.rollback()
            logger.error(f"Error updating segment: {str(e)}")
            raise

    def delete_segment(
        self,
        db: Session,
        group_id: int,
        soft_delete: bool = True
    ) -> bool:
        """
        Delete segment

        Args:
            db: Database session
            group_id: Segment ID
            soft_delete: If True, mark as inactive; if False, remove from database

        Returns:
            True if successful
        """
        try:
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                logger.warning(f"Segment {group_id} not found")
                return False

            if soft_delete:
                # Soft delete: mark as inactive
                segment.is_active = False
                segment.updated_at = datetime.utcnow()
                db.commit()
                logger.info(f"Soft deleted segment {group_id}")
            else:
                # Hard delete: remove members first, then segment
                db.query(ProfileGroupMember).filter(
                    ProfileGroupMember.group_id == group_id
                ).delete()
                db.delete(segment)
                db.commit()
                logger.info(f"Hard deleted segment {group_id}")

            return True

        except Exception as e:
            db.rollback()
            logger.error(f"Error deleting segment: {str(e)}")
            raise

    def refresh_segment(
        self,
        db: Session,
        group_id: int,
        user_id: Optional[int] = None
    ) -> Tuple[int, int]:
        """
        Refresh segment membership based on current filter criteria

        Args:
            db: Database session
            group_id: Segment ID
            user_id: User triggering refresh

        Returns:
            Tuple of (total_members, new_members)
        """
        try:
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                logger.warning(f"Segment {group_id} not found")
                return 0, 0

            # Get current members
            existing_member_ids = set(
                db.query(ProfileGroupMember.profile_id).filter(
                    ProfileGroupMember.group_id == group_id
                ).all()
            )
            existing_member_ids = {m[0] for m in existing_member_ids}

            # Execute filter query to get matching profiles
            matching_profiles = self._execute_filter_query(db, segment)
            matching_profile_ids = {p.profile_id for p in matching_profiles}

            # Add new members
            new_members = matching_profile_ids - existing_member_ids
            for profile_id in new_members:
                member = ProfileGroupMember(
                    group_id=group_id,
                    profile_id=profile_id
                )
                db.add(member)

            # Remove members that no longer match
            removed_members = existing_member_ids - matching_profile_ids
            if removed_members:
                db.query(ProfileGroupMember).filter(
                    and_(
                        ProfileGroupMember.group_id == group_id,
                        ProfileGroupMember.profile_id.in_(removed_members)
                    )
                ).delete(synchronize_session=False)

            # Update segment statistics
            segment.record_count = len(matching_profile_ids)
            segment.last_refreshed = datetime.utcnow()
            segment.last_count_updated = datetime.utcnow()

            # Calculate next refresh time
            if segment.refresh_frequency:
                segment.next_refresh = self._calculate_next_refresh(
                    segment.refresh_frequency
                )

            db.commit()

            # Log refresh
            self._log_query(
                db=db,
                customer_id=segment.customer_id,
                user_id=user_id,
                query_type='SEGMENT',
                filter_query=segment.filter_query,
                result_count=len(matching_profile_ids),
                group_id=group_id,
                group_name=segment.group_name
            )

            logger.info(f"Refreshed segment {group_id}: {len(matching_profile_ids)} members ({len(new_members)} new)")
            return len(matching_profile_ids), len(new_members)

        except Exception as e:
            db.rollback()
            logger.error(f"Error refreshing segment: {str(e)}")
            raise

    def get_segment_members(
        self,
        db: Session,
        group_id: int,
        offset: int = 0,
        limit: int = 100,
        include_profiles: bool = False
    ) -> Tuple[List[Any], int]:
        """
        Get segment members with pagination

        Args:
            db: Database session
            group_id: Segment ID
            offset: Pagination offset
            limit: Results per page
            include_profiles: If True, join and return full Profile objects

        Returns:
            Tuple of (members list, total count)
        """
        try:
            query = db.query(ProfileGroupMember).filter(
                ProfileGroupMember.group_id == group_id
            )

            total = query.count()

            if include_profiles:
                # Join with profiles
                members = db.query(Profile).join(
                    ProfileGroupMember,
                    Profile.profile_id == ProfileGroupMember.profile_id
                ).filter(
                    ProfileGroupMember.group_id == group_id
                ).offset(offset).limit(limit).all()
            else:
                members = query.offset(offset).limit(limit).all()

            logger.info(f"Retrieved {len(members)} members from segment {group_id}")
            return members, total

        except Exception as e:
            logger.error(f"Error fetching segment members: {str(e)}")
            raise

    def add_profiles_to_segment(
        self,
        db: Session,
        group_id: int,
        profile_ids: List[int]
    ) -> int:
        """
        Manually add profiles to a segment (for non-dynamic segments)

        Args:
            db: Database session
            group_id: Segment ID
            profile_ids: List of profile IDs to add

        Returns:
            Number of profiles added
        """
        try:
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                logger.warning(f"Segment {group_id} not found")
                return 0

            if segment.is_dynamic:
                logger.warning(f"Cannot manually add to dynamic segment {group_id}")
                return 0

            count = 0
            for profile_id in profile_ids:
                # Check if already exists
                existing = db.query(ProfileGroupMember).filter(
                    and_(
                        ProfileGroupMember.group_id == group_id,
                        ProfileGroupMember.profile_id == profile_id
                    )
                ).first()

                if not existing:
                    member = ProfileGroupMember(
                        group_id=group_id,
                        profile_id=profile_id
                    )
                    db.add(member)
                    count += 1

            # Update segment count
            segment.record_count = db.query(ProfileGroupMember).filter(
                ProfileGroupMember.group_id == group_id
            ).count()
            segment.last_count_updated = datetime.utcnow()

            db.commit()

            logger.info(f"Added {count} profiles to segment {group_id}")
            return count

        except Exception as e:
            db.rollback()
            logger.error(f"Error adding profiles to segment: {str(e)}")
            raise

    def remove_profiles_from_segment(
        self,
        db: Session,
        group_id: int,
        profile_ids: List[int]
    ) -> int:
        """
        Manually remove profiles from a segment

        Args:
            db: Database session
            group_id: Segment ID
            profile_ids: List of profile IDs to remove

        Returns:
            Number of profiles removed
        """
        try:
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                logger.warning(f"Segment {group_id} not found")
                return 0

            if segment.is_dynamic:
                logger.warning(f"Cannot manually remove from dynamic segment {group_id}")
                return 0

            count = db.query(ProfileGroupMember).filter(
                and_(
                    ProfileGroupMember.group_id == group_id,
                    ProfileGroupMember.profile_id.in_(profile_ids)
                )
            ).delete(synchronize_session=False)

            # Update segment count
            segment.record_count = db.query(ProfileGroupMember).filter(
                ProfileGroupMember.group_id == group_id
            ).count()
            segment.last_count_updated = datetime.utcnow()

            db.commit()

            logger.info(f"Removed {count} profiles from segment {group_id}")
            return count

        except Exception as e:
            db.rollback()
            logger.error(f"Error removing profiles from segment: {str(e)}")
            raise

    def _execute_filter_query(
        self,
        db: Session,
        segment: ProfileGroup
    ) -> List[Profile]:
        """
        Execute segment filter query to get matching profiles

        Args:
            db: Database session
            segment: ProfileGroup object

        Returns:
            List of matching Profile objects
        """
        try:
            # Start with base query
            query = db.query(Profile).filter(
                Profile.customer_id == segment.customer_id
            )

            # Apply filters from filter_query
            filters = segment.filter_query

            if not filters:
                return []

            # Build dynamic query
            query = self._apply_filters_to_query(query, filters)

            profiles = query.all()
            return profiles

        except Exception as e:
            logger.error(f"Error executing filter query: {str(e)}")
            raise

    def _apply_filters_to_query(
        self,
        query,
        filters: Dict[str, Any]
    ):
        """
        Apply filters to SQLAlchemy query

        Args:
            query: SQLAlchemy query object
            filters: Filter dictionary

        Returns:
            Modified query
        """
        try:
            # Handle operator (AND/OR)
            operator = filters.get('operator', 'AND')
            conditions = filters.get('conditions', [])
            groups = filters.get('groups', [])

            filter_list = []

            # Process individual conditions
            for condition in conditions:
                field = condition.get('field')
                op = condition.get('operator')
                value = condition.get('value')

                if not field or not op:
                    continue

                # Standard fields
                if hasattr(Profile, field):
                    column = getattr(Profile, field)

                    if op == 'equals':
                        filter_list.append(column == value)
                    elif op == 'not_equals':
                        filter_list.append(column != value)
                    elif op == 'greater_than':
                        filter_list.append(column > value)
                    elif op == 'greater_than_or_equal':
                        filter_list.append(column >= value)
                    elif op == 'less_than':
                        filter_list.append(column < value)
                    elif op == 'less_than_or_equal':
                        filter_list.append(column <= value)
                    elif op == 'in':
                        filter_list.append(column.in_(value))
                    elif op == 'not_in':
                        filter_list.append(~column.in_(value))
                    elif op == 'contains' and isinstance(value, str):
                        filter_list.append(column.like(f'%{value}%'))
                    elif op == 'starts_with' and isinstance(value, str):
                        filter_list.append(column.like(f'{value}%'))
                    elif op == 'ends_with' and isinstance(value, str):
                        filter_list.append(column.like(f'%{value}'))
                    elif op == 'is_null':
                        filter_list.append(column.is_(None))
                    elif op == 'is_not_null':
                        filter_list.append(column.isnot(None))

                # Custom attributes (JSONB)
                else:
                    if op == 'equals':
                        filter_list.append(
                            Profile.custom_attributes[field].astext == str(value)
                        )
                    elif op == 'contains' and isinstance(value, str):
                        filter_list.append(
                            Profile.custom_attributes[field].astext.like(f'%{value}%')
                        )

            # Process nested groups (recursive)
            for group in groups:
                subquery = self._apply_filters_to_query(query, group)
                # Extract filters from subquery (this is simplified)
                # In practice, you'd need more complex logic here

            # Combine filters
            if filter_list:
                if operator == 'AND':
                    query = query.filter(and_(*filter_list))
                elif operator == 'OR':
                    query = query.filter(or_(*filter_list))

            return query

        except Exception as e:
            logger.error(f"Error applying filters: {str(e)}")
            raise

    def _build_sql_from_filters(
        self,
        filters: Dict[str, Any]
    ) -> str:
        """
        Build SQL WHERE clause from filter dictionary (for display/audit)

        Args:
            filters: Filter dictionary

        Returns:
            SQL WHERE clause string
        """
        try:
            conditions = []
            operator = filters.get('operator', 'AND')

            for condition in filters.get('conditions', []):
                field = condition.get('field')
                op = condition.get('operator')
                value = condition.get('value')

                if op == 'equals':
                    conditions.append(f"{field} = '{value}'")
                elif op == 'not_equals':
                    conditions.append(f"{field} != '{value}'")
                elif op == 'greater_than':
                    conditions.append(f"{field} > {value}")
                elif op == 'less_than':
                    conditions.append(f"{field} < {value}")
                elif op == 'in':
                    values_str = ', '.join([f"'{v}'" for v in value])
                    conditions.append(f"{field} IN ({values_str})")
                elif op == 'contains':
                    conditions.append(f"{field} LIKE '%{value}%'")

            sql = f" {operator} ".join(conditions) if conditions else "1=1"
            return sql

        except Exception as e:
            logger.error(f"Error building SQL: {str(e)}")
            return "ERROR"

    def _generate_group_code(self, db: Session) -> str:
        """Generate unique segment code"""
        import random
        import string

        while True:
            code = 'SEG_' + ''.join(random.choices(string.ascii_uppercase + string.digits, k=12))
            existing = db.query(ProfileGroup).filter(ProfileGroup.group_code == code).first()
            if not existing:
                return code

    def _calculate_next_refresh(self, frequency: str) -> datetime:
        """Calculate next refresh time based on frequency"""
        now = datetime.utcnow()

        if frequency == 'HOURLY':
            return now + timedelta(hours=1)
        elif frequency == 'DAILY':
            return now + timedelta(days=1)
        elif frequency == 'WEEKLY':
            return now + timedelta(weeks=1)
        elif frequency == 'MONTHLY':
            return now + timedelta(days=30)
        else:
            return now + timedelta(days=1)  # Default to daily

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
        """Log query for privacy audit trail"""
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


class QueryBuilderService:
    """Service for building and validating segment queries"""

    def validate_query(
        self,
        query: Dict[str, Any]
    ) -> Tuple[bool, Optional[str]]:
        """
        Validate segment query structure

        Args:
            query: Query dictionary

        Returns:
            Tuple of (is_valid, error_message)
        """
        try:
            # Check required fields
            if 'operator' not in query:
                return False, "Missing 'operator' field"

            if query['operator'] not in ['AND', 'OR']:
                return False, "Operator must be 'AND' or 'OR'"

            # Check conditions
            if 'conditions' in query:
                for condition in query['conditions']:
                    if 'field' not in condition:
                        return False, "Condition missing 'field'"
                    if 'operator' not in condition:
                        return False, "Condition missing 'operator'"

            return True, None

        except Exception as e:
            return False, str(e)

    def build_simple_query(
        self,
        field: str,
        operator: str,
        value: Any
    ) -> Dict[str, Any]:
        """Build a simple single-condition query"""
        return {
            'operator': 'AND',
            'conditions': [
                {
                    'field': field,
                    'operator': operator,
                    'value': value
                }
            ],
            'groups': []
        }

    def build_complex_query(
        self,
        conditions: List[Dict[str, Any]],
        operator: str = 'AND',
        groups: Optional[List[Dict[str, Any]]] = None
    ) -> Dict[str, Any]:
        """Build a complex multi-condition query"""
        return {
            'operator': operator,
            'conditions': conditions,
            'groups': groups or []
        }
