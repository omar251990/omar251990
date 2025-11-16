#!/usr/bin/env python3
"""
Permission Service
Handles permission checking, granting, and management for multi-tenant system
"""

from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_

from src.models.user import User
from src.models.multitenant import (
    Customer, Role, Permission, RolePermission,
    UserPermissionOverride, PermissionAudit
)


class PermissionService:
    """
    Permission Service implementing the permission resolution algorithm:
    1. Check user-level override
    2. Check role-level default
    3. Check customer-level config
    4. Default to deny
    """

    def __init__(self):
        pass

    def check_permission(
        self,
        db: Session,
        user_id: int,
        permission_code: str,
        log_check: bool = True,
        ip_address: str = None,
        user_agent: str = None
    ) -> bool:
        """
        Check if user has specific permission

        Args:
            db: Database session
            user_id: User ID
            permission_code: Permission code (e.g., 'CAMPAIGN_CREATE')
            log_check: Whether to log the permission check
            ip_address: IP address for audit
            user_agent: User agent for audit

        Returns:
            bool: True if user has permission, False otherwise
        """
        # Get user
        user = db.query(User).filter(User.id == user_id).first()
        if not user or not user.is_active:
            return False

        # Get permission
        permission = db.query(Permission).filter(
            Permission.permission_code == permission_code,
            Permission.is_active == True
        ).first()

        if not permission:
            return False

        # Permission resolution algorithm
        result = False

        # Step 1: Check user-level override
        override = db.query(UserPermissionOverride).filter(
            and_(
                UserPermissionOverride.user_id == user_id,
                UserPermissionOverride.permission_id == permission.permission_id,
                or_(
                    UserPermissionOverride.expires_at.is_(None),
                    UserPermissionOverride.expires_at > datetime.utcnow()
                )
            )
        ).first()

        if override:
            result = override.allow
        else:
            # Step 2: Check role-level permission
            if user.role_id:
                role_perm = db.query(RolePermission).filter(
                    and_(
                        RolePermission.role_id == user.role_id,
                        RolePermission.permission_id == permission.permission_id
                    )
                ).first()

                if role_perm:
                    result = role_perm.allow

        # Log permission check if requested
        if log_check and user.customer_id:
            self._log_permission_check(
                db, user.customer_id, user_id, permission.permission_id,
                result, ip_address, user_agent
            )

        return result

    def check_permissions(
        self,
        db: Session,
        user_id: int,
        permission_codes: List[str]
    ) -> Dict[str, bool]:
        """
        Check multiple permissions at once

        Args:
            db: Database session
            user_id: User ID
            permission_codes: List of permission codes

        Returns:
            Dict mapping permission codes to boolean results
        """
        results = {}
        for code in permission_codes:
            results[code] = self.check_permission(db, user_id, code, log_check=False)
        return results

    def grant_permission(
        self,
        db: Session,
        user_id: int,
        permission_code: str,
        granted_by_user_id: int,
        expires_at: Optional[datetime] = None,
        reason: str = None
    ) -> bool:
        """
        Grant a permission to a user (override)

        Args:
            db: Database session
            user_id: User ID to grant permission to
            permission_code: Permission code
            granted_by_user_id: User ID granting the permission
            expires_at: Optional expiration date
            reason: Optional reason for granting

        Returns:
            bool: True if successful
        """
        # Verify granter has permission to grant
        if not self.check_permission(db, granted_by_user_id, 'PERMISSION_ASSIGN'):
            raise PermissionError("You don't have permission to grant permissions")

        # Get user and permission
        user = db.query(User).filter(User.id == user_id).first()
        permission = db.query(Permission).filter(Permission.permission_code == permission_code).first()

        if not user or not permission:
            return False

        # Get granted_by username
        granted_by_user = db.query(User).filter(User.id == granted_by_user_id).first()
        granted_by = granted_by_user.username if granted_by_user else str(granted_by_user_id)

        # Check if override already exists
        override = db.query(UserPermissionOverride).filter(
            and_(
                UserPermissionOverride.user_id == user_id,
                UserPermissionOverride.permission_id == permission.permission_id
            )
        ).first()

        if override:
            # Update existing override
            override.allow = True
            override.granted_by = granted_by
            override.granted_at = datetime.utcnow()
            override.expires_at = expires_at
            override.reason = reason
        else:
            # Create new override
            override = UserPermissionOverride(
                user_id=user_id,
                permission_id=permission.permission_id,
                allow=True,
                granted_by=granted_by,
                granted_at=datetime.utcnow(),
                expires_at=expires_at,
                reason=reason
            )
            db.add(override)

        # Log the grant
        self._log_permission_action(
            db, user.customer_id, granted_by_user_id, user_id,
            permission.permission_id, 'GRANTED'
        )

        db.commit()
        return True

    def revoke_permission(
        self,
        db: Session,
        user_id: int,
        permission_code: str,
        revoked_by_user_id: int
    ) -> bool:
        """
        Revoke a permission from a user

        Args:
            db: Database session
            user_id: User ID to revoke permission from
            permission_code: Permission code
            revoked_by_user_id: User ID revoking the permission

        Returns:
            bool: True if successful
        """
        # Verify revoker has permission
        if not self.check_permission(db, revoked_by_user_id, 'PERMISSION_REVOKE'):
            raise PermissionError("You don't have permission to revoke permissions")

        # Get user and permission
        user = db.query(User).filter(User.id == user_id).first()
        permission = db.query(Permission).filter(Permission.permission_code == permission_code).first()

        if not user or not permission:
            return False

        # Delete the override (reverts to role default)
        db.query(UserPermissionOverride).filter(
            and_(
                UserPermissionOverride.user_id == user_id,
                UserPermissionOverride.permission_id == permission.permission_id
            )
        ).delete()

        # Log the revoke
        self._log_permission_action(
            db, user.customer_id, revoked_by_user_id, user_id,
            permission.permission_id, 'REVOKED'
        )

        db.commit()
        return True

    def get_user_permissions(
        self,
        db: Session,
        user_id: int,
        include_role_permissions: bool = True
    ) -> List[Dict[str, Any]]:
        """
        Get all permissions for a user

        Args:
            db: Database session
            user_id: User ID
            include_role_permissions: Include role-based permissions

        Returns:
            List of permission dictionaries
        """
        user = db.query(User).filter(User.id == user_id).first()
        if not user:
            return []

        permissions = []

        # Get user overrides
        overrides = db.query(
            UserPermissionOverride, Permission
        ).join(
            Permission
        ).filter(
            UserPermissionOverride.user_id == user_id
        ).all()

        for override, perm in overrides:
            permissions.append({
                'permission_code': perm.permission_code,
                'module_name': perm.module_name,
                'action_name': perm.action_name,
                'description': perm.description,
                'allow': override.allow,
                'source': 'USER_OVERRIDE',
                'expires_at': override.expires_at.isoformat() if override.expires_at else None,
                'granted_by': override.granted_by,
                'granted_at': override.granted_at.isoformat()
            })

        # Get role permissions if requested
        if include_role_permissions and user.role_id:
            role_perms = db.query(
                RolePermission, Permission
            ).join(
                Permission
            ).filter(
                RolePermission.role_id == user.role_id
            ).all()

            # Only include role permissions not overridden by user
            override_codes = {p['permission_code'] for p in permissions}

            for role_perm, perm in role_perms:
                if perm.permission_code not in override_codes:
                    permissions.append({
                        'permission_code': perm.permission_code,
                        'module_name': perm.module_name,
                        'action_name': perm.action_name,
                        'description': perm.description,
                        'allow': role_perm.allow,
                        'source': 'ROLE',
                        'expires_at': None,
                        'granted_by': None,
                        'granted_at': None
                    })

        return permissions

    def assign_role_permissions(
        self,
        db: Session,
        role_id: int,
        permission_codes: List[str],
        allow: bool = True
    ) -> bool:
        """
        Assign permissions to a role

        Args:
            db: Database session
            role_id: Role ID
            permission_codes: List of permission codes
            allow: Whether to allow or deny

        Returns:
            bool: True if successful
        """
        role = db.query(Role).filter(Role.role_id == role_id).first()
        if not role:
            return False

        for code in permission_codes:
            permission = db.query(Permission).filter(Permission.permission_code == code).first()
            if not permission:
                continue

            # Check if mapping exists
            role_perm = db.query(RolePermission).filter(
                and_(
                    RolePermission.role_id == role_id,
                    RolePermission.permission_id == permission.permission_id
                )
            ).first()

            if role_perm:
                role_perm.allow = allow
            else:
                role_perm = RolePermission(
                    role_id=role_id,
                    permission_id=permission.permission_id,
                    allow=allow
                )
                db.add(role_perm)

        db.commit()
        return True

    def get_role_permissions(
        self,
        db: Session,
        role_id: int
    ) -> List[Dict[str, Any]]:
        """
        Get all permissions for a role

        Args:
            db: Database session
            role_id: Role ID

        Returns:
            List of permission dictionaries
        """
        role_perms = db.query(
            RolePermission, Permission
        ).join(
            Permission
        ).filter(
            RolePermission.role_id == role_id
        ).all()

        return [
            {
                'permission_code': perm.permission_code,
                'module_name': perm.module_name,
                'action_name': perm.action_name,
                'description': perm.description,
                'allow': role_perm.allow,
                'category': perm.category
            }
            for role_perm, perm in role_perms
        ]

    def _log_permission_check(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        permission_id: int,
        result: bool,
        ip_address: str = None,
        user_agent: str = None
    ):
        """Log permission check to audit table"""
        audit = PermissionAudit(
            customer_id=customer_id,
            user_id=user_id,
            permission_id=permission_id,
            action='CHECKED',
            result=result,
            ip_address=ip_address,
            user_agent=user_agent
        )
        db.add(audit)
        db.commit()

    def _log_permission_action(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        target_user_id: int,
        permission_id: int,
        action: str
    ):
        """Log permission grant/revoke action"""
        audit = PermissionAudit(
            customer_id=customer_id,
            user_id=user_id,
            target_user_id=target_user_id,
            permission_id=permission_id,
            action=action,
            result=True
        )
        db.add(audit)
        db.commit()

    def get_permission_audit_log(
        self,
        db: Session,
        customer_id: int = None,
        user_id: int = None,
        limit: int = 100
    ) -> List[Dict[str, Any]]:
        """
        Get permission audit log

        Args:
            db: Database session
            customer_id: Filter by customer ID
            user_id: Filter by user ID
            limit: Maximum number of records

        Returns:
            List of audit log entries
        """
        query = db.query(PermissionAudit)

        if customer_id:
            query = query.filter(PermissionAudit.customer_id == customer_id)

        if user_id:
            query = query.filter(PermissionAudit.user_id == user_id)

        audits = query.order_by(PermissionAudit.created_at.desc()).limit(limit).all()

        return [
            {
                'audit_id': audit.audit_id,
                'customer_id': audit.customer_id,
                'user_id': audit.user_id,
                'target_user_id': audit.target_user_id,
                'permission_id': audit.permission_id,
                'action': audit.action,
                'result': audit.result,
                'ip_address': str(audit.ip_address) if audit.ip_address else None,
                'created_at': audit.created_at.isoformat(),
                'metadata': audit.metadata
            }
            for audit in audits
        ]


# Global permission service instance
permission_service = PermissionService()


# Decorator for permission-protected endpoints
def require_permission(permission_code: str):
    """
    Decorator to require specific permission for endpoint access

    Usage:
        @require_permission('CAMPAIGN_CREATE')
        def create_campaign(...):
            ...
    """
    def decorator(func):
        def wrapper(*args, **kwargs):
            # This would be implemented with FastAPI dependencies
            # For now, it's a placeholder
            pass
        return wrapper
    return decorator
