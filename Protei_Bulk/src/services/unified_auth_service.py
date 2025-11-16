#!/usr/bin/env python3
"""
Unified Authentication Service
Handles authentication for Web Portal, HTTP API Gateway, and SMPP Gateway
"""

from typing import Optional, Dict, Any, Tuple
from datetime import datetime, timedelta
import secrets
import hashlib
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_, func

from src.models.user import User
from src.services.auth import AuthService


class UnifiedAuthService:
    """
    Unified Authentication Service for all access channels:
    - Web Portal (Username + Password + 2FA)
    - HTTP API (API Key or Basic Auth)
    - SMPP Gateway (System ID + Password)
    """

    def __init__(self):
        self.auth_service = AuthService()

    # ============================================================================
    # WEB PORTAL AUTHENTICATION
    # ============================================================================

    def authenticate_web(
        self,
        db: Session,
        username: str,
        password: str,
        otp_code: str = None,
        ip_address: str = None,
        user_agent: str = None
    ) -> Tuple[bool, Optional[User], Optional[str], Optional[Dict]]:
        """
        Authenticate user for web portal access

        Args:
            db: Database session
            username: Username or email
            password: User password
            otp_code: 2FA OTP code (if enabled)
            ip_address: Client IP address
            user_agent: Client user agent

        Returns:
            Tuple: (success, user, jwt_token, error_dict)
        """
        # Verify credentials
        user = self.auth_service.verify_credentials(db, username, password)
        if not user:
            self._log_auth_attempt(db, None, 'WEB', False, ip_address, 'Invalid credentials')
            return False, None, None, {'error': 'Invalid username or password'}

        # Check if user is active
        if not user.is_active:
            self._log_auth_attempt(db, user.id, 'WEB', False, ip_address, 'User inactive')
            return False, None, None, {'error': 'Account is inactive'}

        # Check if user is allowed to use web portal
        if hasattr(user, 'bind_type') and user.bind_type == 'SMPP':
            self._log_auth_attempt(db, user.id, 'WEB', False, ip_address, 'Web access not allowed')
            return False, None, None, {'error': 'Web portal access not allowed for this account'}

        # Check 2FA if enabled
        if user.two_factor_enabled:
            if not otp_code:
                return False, None, None, {'error': '2FA code required', 'require_2fa': True}

            if not self.auth_service.verify_totp(user.two_factor_secret, otp_code):
                self._log_auth_attempt(db, user.id, 'WEB', False, ip_address, 'Invalid 2FA code')
                return False, None, None, {'error': 'Invalid 2FA code'}

        # Check customer status
        if hasattr(user, 'customer') and user.customer:
            if user.customer.status != 'ACTIVE':
                self._log_auth_attempt(db, user.id, 'WEB', False, ip_address, 'Customer suspended')
                return False, None, None, {'error': 'Your account has been suspended. Please contact support.'}

            if user.customer.expiry_date and user.customer.expiry_date < datetime.utcnow():
                self._log_auth_attempt(db, user.id, 'WEB', False, ip_address, 'License expired')
                return False, None, None, {'error': 'Your license has expired. Please renew to continue.'}

        # Generate JWT token
        jwt_token = self.auth_service.create_access_token(user.user_id)

        # Update last login
        user.last_login = datetime.utcnow()
        user.last_ip = ip_address
        db.commit()

        # Log successful authentication
        self._log_auth_attempt(db, user.id, 'WEB', True, ip_address, 'Success')

        return True, user, jwt_token, None

    # ============================================================================
    # HTTP API AUTHENTICATION
    # ============================================================================

    def authenticate_api_key(
        self,
        db: Session,
        api_key: str,
        ip_address: str = None
    ) -> Tuple[bool, Optional[User], Optional[Dict]]:
        """
        Authenticate user using API key

        Args:
            db: Database session
            api_key: API key from Authorization header
            ip_address: Client IP address

        Returns:
            Tuple: (success, user, error_dict)
        """
        if not api_key:
            return False, None, {'error': 'API key required'}

        # Find user by API key
        user = db.query(User).filter(
            and_(
                User.api_key == api_key,
                User.is_active == True
            )
        ).first()

        if not user:
            self._log_auth_attempt(db, None, 'HTTP_API', False, ip_address, 'Invalid API key')
            return False, None, {'error': 'Invalid API key'}

        # Check if user is allowed to use HTTP API
        if hasattr(user, 'can_use_http') and not user.can_use_http:
            self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'HTTP API access not allowed')
            return False, None, {'error': 'HTTP API access not allowed for this account'}

        if hasattr(user, 'can_use_api_bulk') and not user.can_use_api_bulk:
            self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'API bulk access not allowed')
            return False, None, {'error': 'API bulk access not allowed for this account'}

        # Check customer status
        if hasattr(user, 'customer') and user.customer:
            if user.customer.status != 'ACTIVE':
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'Customer suspended')
                return False, None, {'error': 'Account suspended'}

            if user.customer.expiry_date and user.customer.expiry_date < datetime.utcnow():
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'License expired')
                return False, None, {'error': 'License expired'}

        # Check daily quota
        if hasattr(user, 'max_msg_per_day'):
            if not self._check_daily_quota(db, user.id, user.max_msg_per_day):
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'Daily quota exceeded')
                return False, None, {'error': 'Daily quota exceeded'}

        # Update last activity
        user.last_login = datetime.utcnow()
        user.last_ip = ip_address
        db.commit()

        # Log successful authentication
        self._log_auth_attempt(db, user.id, 'HTTP_API', True, ip_address, 'Success')

        return True, user, None

    def authenticate_basic_auth(
        self,
        db: Session,
        username: str,
        password: str,
        ip_address: str = None
    ) -> Tuple[bool, Optional[User], Optional[Dict]]:
        """
        Authenticate user using Basic Authentication

        Args:
            db: Database session
            username: Username
            password: Password
            ip_address: Client IP address

        Returns:
            Tuple: (success, user, error_dict)
        """
        # Verify credentials
        user = self.auth_service.verify_credentials(db, username, password)
        if not user:
            self._log_auth_attempt(db, None, 'HTTP_API', False, ip_address, 'Invalid credentials')
            return False, None, {'error': 'Invalid username or password'}

        # Check if user is active
        if not user.is_active:
            self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'User inactive')
            return False, None, {'error': 'Account is inactive'}

        # Check if user is allowed to use HTTP API
        if hasattr(user, 'can_use_http') and not user.can_use_http:
            self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'HTTP API access not allowed')
            return False, None, {'error': 'HTTP API access not allowed for this account'}

        # Check customer status
        if hasattr(user, 'customer') and user.customer:
            if user.customer.status != 'ACTIVE':
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'Customer suspended')
                return False, None, {'error': 'Account suspended'}

            if user.customer.expiry_date and user.customer.expiry_date < datetime.utcnow():
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'License expired')
                return False, None, {'error': 'License expired'}

        # Check daily quota
        if hasattr(user, 'max_msg_per_day'):
            if not self._check_daily_quota(db, user.id, user.max_msg_per_day):
                self._log_auth_attempt(db, user.id, 'HTTP_API', False, ip_address, 'Daily quota exceeded')
                return False, None, {'error': 'Daily quota exceeded'}

        # Update last activity
        user.last_login = datetime.utcnow()
        user.last_ip = ip_address
        db.commit()

        # Log successful authentication
        self._log_auth_attempt(db, user.id, 'HTTP_API', True, ip_address, 'Success')

        return True, user, None

    # ============================================================================
    # SMPP AUTHENTICATION
    # ============================================================================

    def authenticate_smpp(
        self,
        db: Session,
        system_id: str,
        password: str,
        bind_type: str,
        remote_ip: str,
        remote_port: int = None
    ) -> Tuple[bool, Optional[User], Optional[str], Optional[Dict]]:
        """
        Authenticate SMPP bind request

        Args:
            db: Database session
            system_id: SMPP system_id (username)
            password: SMPP password
            bind_type: TRANSMITTER, RECEIVER, or TRANSCEIVER
            remote_ip: Client IP address
            remote_port: Client port

        Returns:
            Tuple: (success, user, session_token, error_dict)
        """
        # Verify credentials
        user = self.auth_service.verify_credentials(db, system_id, password)
        if not user:
            self._log_auth_attempt(db, None, 'SMPP', False, remote_ip, 'Invalid credentials')
            return False, None, None, {'error': 'Invalid system_id or password'}

        # Check if user is active
        if not user.is_active:
            self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'User inactive')
            return False, None, None, {'error': 'Account is inactive'}

        # Check if user is allowed to use SMPP
        if hasattr(user, 'can_use_smpp') and not user.can_use_smpp:
            self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'SMPP access not allowed')
            return False, None, None, {'error': 'SMPP access not allowed for this account'}

        # Check bind type compatibility
        if hasattr(user, 'bind_type'):
            allowed_bind_types = user.bind_type
            if allowed_bind_types == 'HTTP' or allowed_bind_types == 'WEB_ONLY':
                self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'SMPP bind not allowed')
                return False, None, None, {'error': 'SMPP bind not allowed for this account'}

        # Check customer status
        if hasattr(user, 'customer') and user.customer:
            if user.customer.status != 'ACTIVE':
                self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'Customer suspended')
                return False, None, None, {'error': 'Account suspended'}

            if user.customer.expiry_date and user.customer.expiry_date < datetime.utcnow():
                self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'License expired')
                return False, None, None, {'error': 'License expired'}

        # Check daily quota
        if hasattr(user, 'max_msg_per_day'):
            if not self._check_daily_quota(db, user.id, user.max_msg_per_day):
                self._log_auth_attempt(db, user.id, 'SMPP', False, remote_ip, 'Daily quota exceeded')
                return False, None, None, {'error': 'Daily quota exceeded'}

        # Generate session token
        session_token = secrets.token_hex(32)

        # Create SMPP session record
        from src.models.smpp import SMPPSession
        smpp_session = SMPPSession(
            user_id=user.id,
            customer_id=user.customer_id if hasattr(user, 'customer_id') else None,
            system_id=system_id,
            bind_type=bind_type,
            remote_ip=remote_ip,
            remote_port=remote_port,
            session_token=session_token,
            status='BOUND',
            bound_at=datetime.utcnow(),
            last_activity_at=datetime.utcnow()
        )
        db.add(smpp_session)

        # Update user last login
        user.last_login = datetime.utcnow()
        user.last_ip = remote_ip
        db.commit()

        # Log successful authentication
        self._log_auth_attempt(db, user.id, 'SMPP', True, remote_ip, f'Bind {bind_type} success')

        return True, user, session_token, None

    def disconnect_smpp(
        self,
        db: Session,
        session_token: str
    ) -> bool:
        """
        Disconnect SMPP session

        Args:
            db: Database session
            session_token: Session token

        Returns:
            bool: Success
        """
        from src.models.smpp import SMPPSession

        session = db.query(SMPPSession).filter(
            SMPPSession.session_token == session_token
        ).first()

        if session:
            session.status = 'DISCONNECTED'
            session.disconnected_at = datetime.utcnow()
            db.commit()
            return True

        return False

    # ============================================================================
    # API KEY MANAGEMENT
    # ============================================================================

    def generate_api_key(
        self,
        db: Session,
        user_id: int
    ) -> Optional[str]:
        """
        Generate new API key for user

        Args:
            db: Database session
            user_id: User ID

        Returns:
            str: Generated API key
        """
        user = db.query(User).filter(User.id == user_id).first()
        if not user:
            return None

        # Generate secure API key
        api_key = secrets.token_urlsafe(48)

        # Update user record
        user.api_key = api_key
        user.api_key_created_at = datetime.utcnow()
        db.commit()

        return api_key

    def revoke_api_key(
        self,
        db: Session,
        user_id: int
    ) -> bool:
        """
        Revoke API key for user

        Args:
            db: Database session
            user_id: User ID

        Returns:
            bool: Success
        """
        user = db.query(User).filter(User.id == user_id).first()
        if not user:
            return False

        user.api_key = None
        user.api_key_created_at = None
        db.commit()

        return True

    # ============================================================================
    # QUOTA CHECKING
    # ============================================================================

    def _check_daily_quota(
        self,
        db: Session,
        user_id: int,
        max_msg_per_day: int
    ) -> bool:
        """
        Check if user has not exceeded daily quota

        Args:
            db: Database session
            user_id: User ID
            max_msg_per_day: Maximum messages per day

        Returns:
            bool: True if under quota
        """
        from src.models.quota import QuotaUsage

        today = datetime.utcnow().date()

        usage = db.query(QuotaUsage).filter(
            and_(
                QuotaUsage.user_id == user_id,
                QuotaUsage.period_date == today,
                QuotaUsage.period_type == 'DAILY'
            )
        ).first()

        if not usage:
            return True

        return usage.messages_sent < max_msg_per_day

    def get_daily_usage(
        self,
        db: Session,
        user_id: int
    ) -> Dict[str, Any]:
        """
        Get daily usage statistics for user

        Args:
            db: Database session
            user_id: User ID

        Returns:
            Dict: Usage statistics
        """
        from src.models.quota import QuotaUsage

        today = datetime.utcnow().date()

        usage = db.query(QuotaUsage).filter(
            and_(
                QuotaUsage.user_id == user_id,
                QuotaUsage.period_date == today,
                QuotaUsage.period_type == 'DAILY'
            )
        ).first()

        user = db.query(User).filter(User.id == user_id).first()
        max_messages = user.max_msg_per_day if hasattr(user, 'max_msg_per_day') else 500000

        if not usage:
            return {
                'messages_sent': 0,
                'messages_delivered': 0,
                'messages_failed': 0,
                'max_messages': max_messages,
                'remaining': max_messages,
                'percentage_used': 0.0
            }

        remaining = max(0, max_messages - usage.messages_sent)
        percentage = (usage.messages_sent / max_messages * 100) if max_messages > 0 else 0

        return {
            'messages_sent': usage.messages_sent,
            'messages_delivered': usage.messages_delivered,
            'messages_failed': usage.messages_failed,
            'max_messages': max_messages,
            'remaining': remaining,
            'percentage_used': round(percentage, 2)
        }

    # ============================================================================
    # AUDIT LOGGING
    # ============================================================================

    def _log_auth_attempt(
        self,
        db: Session,
        user_id: Optional[int],
        channel: str,
        success: bool,
        ip_address: str = None,
        details: str = None
    ):
        """
        Log authentication attempt

        Args:
            db: Database session
            user_id: User ID (if known)
            channel: Authentication channel (WEB, HTTP_API, SMPP)
            success: Whether authentication succeeded
            ip_address: Client IP address
            details: Additional details
        """
        from src.models.audit import AuditLog

        try:
            log = AuditLog(
                user_id=user_id,
                action='AUTH_ATTEMPT',
                entity_type='USER',
                entity_id=user_id,
                ip_address=ip_address,
                details=f"Channel: {channel}, Success: {success}, Details: {details}",
                metadata={
                    'channel': channel,
                    'success': success,
                    'details': details
                }
            )
            db.add(log)
            db.commit()
        except Exception as e:
            # Don't fail authentication due to logging error
            db.rollback()


# Global unified auth service instance
unified_auth_service = UnifiedAuthService()
