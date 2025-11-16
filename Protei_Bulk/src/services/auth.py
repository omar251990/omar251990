"""
Authentication Service
Handles user authentication, JWT tokens, 2FA, and API keys
"""

from datetime import datetime, timedelta
from typing import Optional, Dict
from passlib.context import CryptContext
from jose import JWTError, jwt
import pyotp
import secrets

from src.core.config import get_config
from src.models.user import User
from sqlalchemy.orm import Session

# Password hashing
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


class AuthService:
    """Authentication service"""

    def __init__(self):
        self.config = get_config()

    def hash_password(self, password: str) -> str:
        """Hash a password"""
        return pwd_context.hash(password)

    def verify_password(self, plain_password: str, hashed_password: str) -> bool:
        """Verify a password against its hash"""
        return pwd_context.verify(plain_password, hashed_password)

    def create_access_token(self, data: dict, expires_delta: Optional[timedelta] = None) -> str:
        """Create a JWT access token"""
        to_encode = data.copy()

        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(
                minutes=self.config.security.access_token_expire_minutes
            )

        to_encode.update({"exp": expire, "type": "access"})
        encoded_jwt = jwt.encode(
            to_encode,
            self.config.security.secret_key,
            algorithm=self.config.security.algorithm
        )
        return encoded_jwt

    def create_refresh_token(self, data: dict) -> str:
        """Create a JWT refresh token"""
        to_encode = data.copy()
        expire = datetime.utcnow() + timedelta(days=self.config.security.refresh_token_expire_days)
        to_encode.update({"exp": expire, "type": "refresh"})
        encoded_jwt = jwt.encode(
            to_encode,
            self.config.security.secret_key,
            algorithm=self.config.security.algorithm
        )
        return encoded_jwt

    def verify_token(self, token: str) -> Optional[Dict]:
        """Verify and decode a JWT token"""
        try:
            payload = jwt.decode(
                token,
                self.config.security.secret_key,
                algorithms=[self.config.security.algorithm]
            )
            return payload
        except JWTError:
            return None

    def generate_api_key(self) -> str:
        """Generate a new API key"""
        return secrets.token_urlsafe(32)

    def generate_2fa_secret(self) -> str:
        """Generate a new 2FA secret"""
        return pyotp.random_base32()

    def verify_2fa_token(self, secret: str, token: str) -> bool:
        """Verify a 2FA token"""
        totp = pyotp.TOTP(secret)
        return totp.verify(token, valid_window=1)

    def get_2fa_qr_code_url(self, username: str, secret: str) -> str:
        """Get 2FA QR code provisioning URL"""
        totp = pyotp.TOTP(secret)
        return totp.provisioning_uri(
            name=username,
            issuer_name=self.config.app.app_name
        )

    def authenticate_user(self, db: Session, username: str, password: str) -> Optional[User]:
        """Authenticate a user with username and password"""
        user = db.query(User).filter(User.username == username).first()

        if not user:
            return None

        # Check if account is locked
        if user.locked_until and user.locked_until > datetime.utcnow():
            return None

        # Check if account is active
        if user.status != "ACTIVE":
            return None

        # Verify password
        if not self.verify_password(password, user.password_hash):
            # Increment failed login attempts
            user.failed_login_attempts += 1

            # Lock account if too many failures
            if user.failed_login_attempts >= self.config.security.max_failed_attempts:
                user.locked_until = datetime.utcnow() + timedelta(
                    minutes=self.config.security.lockout_duration_minutes
                )

            db.commit()
            return None

        # Reset failed attempts on successful login
        user.failed_login_attempts = 0
        user.last_login_at = datetime.utcnow()
        db.commit()

        return user

    def verify_api_key(self, db: Session, api_key: str) -> Optional[User]:
        """Verify an API key and return the associated user"""
        user = db.query(User).filter(User.api_key == api_key, User.api_enabled == True).first()

        if not user:
            return None

        # Check if API key expired
        if user.api_key_expires_at and user.api_key_expires_at < datetime.utcnow():
            return None

        # Check if user is active
        if user.status != "ACTIVE":
            return None

        return user


# Global auth service instance
_auth_service = None


def get_auth_service() -> AuthService:
    """Get global auth service instance"""
    global _auth_service
    if _auth_service is None:
        _auth_service = AuthService()
    return _auth_service
