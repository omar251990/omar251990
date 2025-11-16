"""
Configuration Management System
Loads and manages all application configuration from config files
"""

import os
import configparser
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass


@dataclass
class DatabaseConfig:
    """Database configuration"""
    host: str = "localhost"
    port: int = 5432
    database: str = "protei_bulk"
    username: str = "protei"
    password: str = "elephant"
    ssl_mode: str = "require"
    pool_size: int = 20
    max_connections: int = 50

    @property
    def url(self) -> str:
        """Get SQLAlchemy database URL"""
        return f"postgresql://{self.username}:{self.password}@{self.host}:{self.port}/{self.database}"

    @property
    def async_url(self) -> str:
        """Get async SQLAlchemy database URL"""
        return f"postgresql+asyncpg://{self.username}:{self.password}@{self.host}:{self.port}/{self.database}"


@dataclass
class RedisConfig:
    """Redis configuration"""
    enabled: bool = True
    host: str = "localhost"
    port: int = 6379
    password: str = ""
    database: int = 0
    pool_size: int = 10

    @property
    def url(self) -> str:
        """Get Redis connection URL"""
        if self.password:
            return f"redis://:{self.password}@{self.host}:{self.port}/{self.database}"
        return f"redis://{self.host}:{self.port}/{self.database}"


@dataclass
class AppConfig:
    """Application configuration"""
    app_name: str = "Protei_Bulk"
    version: str = "1.0.0"
    build: str = "001"
    environment: str = "production"
    base_dir: Path = Path("/opt/Protei_Bulk")
    max_workers: int = 10
    queue_size: int = 10000
    enable_monitoring: bool = True


@dataclass
class SMPPConfig:
    """SMPP protocol configuration"""
    enabled: bool = True
    bind_address: str = "0.0.0.0"
    bind_port: int = 2775
    system_id: str = "PROTEI_BULK"
    max_connections: int = 100
    enquire_link_interval: int = 30


@dataclass
class APIConfig:
    """API configuration"""
    enabled: bool = True
    bind_address: str = "0.0.0.0"
    bind_port: int = 8080
    enable_https: bool = False
    api_base_path: str = "/api/v1"
    enable_cors: bool = True
    cors_origins: list = None

    def __post_init__(self):
        if self.cors_origins is None:
            self.cors_origins = ["*"]


@dataclass
class SecurityConfig:
    """Security configuration"""
    secret_key: str = ""  # Will be generated
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 60
    refresh_token_expire_days: int = 7
    password_min_length: int = 12
    password_expiry_days: int = 90
    max_failed_attempts: int = 5
    lockout_duration_minutes: int = 30
    enable_2fa: bool = True


class Config:
    """Main configuration class"""

    def __init__(self, base_dir: Optional[Path] = None):
        if base_dir is None:
            # Try to detect base directory
            current = Path(__file__).resolve()
            # Go up from src/core/config.py to Protei_Bulk/
            base_dir = current.parent.parent.parent

        self.base_dir = Path(base_dir)
        self.config_dir = self.base_dir / "config"

        # Initialize all configs
        self.app = AppConfig(base_dir=self.base_dir)
        self.database = DatabaseConfig()
        self.redis = RedisConfig()
        self.smpp = SMPPConfig()
        self.api = APIConfig()
        self.security = SecurityConfig()

        # Load from config files
        self._load_configs()

        # Generate secret key if not set
        if not self.security.secret_key:
            import secrets
            self.security.secret_key = secrets.token_urlsafe(32)

    def _load_configs(self):
        """Load all configuration files"""
        self._load_app_config()
        self._load_db_config()
        self._load_protocol_config()
        self._load_security_config()

    def _load_app_config(self):
        """Load app.conf"""
        config_file = self.config_dir / "app.conf"
        if not config_file.exists():
            return

        parser = configparser.ConfigParser()
        parser.read(config_file)

        if "Application" in parser:
            app = parser["Application"]
            self.app.app_name = app.get("app_name", self.app.app_name)
            self.app.version = app.get("version", self.app.version)
            self.app.build = app.get("build", self.app.build)
            self.app.environment = app.get("environment", self.app.environment)

        if "Runtime" in parser:
            runtime = parser["Runtime"]
            self.app.max_workers = runtime.getint("max_workers", self.app.max_workers)
            self.app.queue_size = runtime.getint("queue_size", self.app.queue_size)

        if "Performance" in parser:
            perf = parser["Performance"]
            self.app.enable_monitoring = perf.getboolean("enable_monitoring", self.app.enable_monitoring)

    def _load_db_config(self):
        """Load db.conf"""
        # First, load from environment variables (Docker)
        if os.getenv("DB_HOST"):
            self.database.host = os.getenv("DB_HOST", self.database.host)
            self.database.port = int(os.getenv("DB_PORT", self.database.port))
            self.database.database = os.getenv("DB_NAME", self.database.database)
            self.database.username = os.getenv("DB_USER", self.database.username)
            self.database.password = os.getenv("DB_PASSWORD", self.database.password)

        if os.getenv("REDIS_HOST"):
            self.redis.host = os.getenv("REDIS_HOST", self.redis.host)
            self.redis.port = int(os.getenv("REDIS_PORT", self.redis.port))
            self.redis.password = os.getenv("REDIS_PASSWORD", self.redis.password)
            redis_db = os.getenv("REDIS_DB", "0")
            self.redis.database = int(redis_db) if redis_db else 0

        # Then, try to load from config file (optional)
        config_file = self.config_dir / "db.conf"
        if not config_file.exists():
            return

        try:
            parser = configparser.ConfigParser()
            parser.read(config_file)

            if "PostgreSQL" in parser:
                db = parser["PostgreSQL"]
                # Only override if env vars not set
                if not os.getenv("DB_HOST"):
                    self.database.host = db.get("host", self.database.host)
                    self.database.port = db.getint("port", self.database.port)
                    self.database.database = db.get("database", self.database.database)
                    self.database.username = db.get("username", self.database.username)
                    self.database.password = db.get("password", self.database.password)
                    self.database.pool_size = db.getint("pool_size", self.database.pool_size)

            if "Redis" in parser:
                redis = parser["Redis"]
                # Only override if env vars not set
                if not os.getenv("REDIS_HOST"):
                    self.redis.enabled = redis.getboolean("enabled", self.redis.enabled)
                    self.redis.host = redis.get("host", self.redis.host)
                    self.redis.port = redis.getint("port", self.redis.port)
                    self.redis.password = redis.get("password", self.redis.password)
                    self.redis.database = redis.getint("database", self.redis.database)
        except Exception as e:
            # If config file is malformed, just use env vars or defaults
            import logging
            logging.warning(f"Failed to load db.conf: {e}. Using environment variables or defaults.")

    def _load_protocol_config(self):
        """Load protocol.conf"""
        config_file = self.config_dir / "protocol.conf"
        if not config_file.exists():
            return

        parser = configparser.ConfigParser()
        parser.read(config_file)

        if "SMPP" in parser:
            smpp = parser["SMPP"]
            self.smpp.enabled = smpp.getboolean("enabled", self.smpp.enabled)
            self.smpp.bind_address = smpp.get("bind_address", self.smpp.bind_address)
            self.smpp.bind_port = smpp.getint("bind_port", self.smpp.bind_port)
            self.smpp.system_id = smpp.get("system_id", self.smpp.system_id)
            self.smpp.max_connections = smpp.getint("max_connections", self.smpp.max_connections)

        if "HTTP" in parser:
            http = parser["HTTP"]
            self.api.enabled = http.getboolean("enabled", self.api.enabled)
            self.api.bind_address = http.get("bind_address", self.api.bind_address)
            self.api.bind_port = http.getint("bind_port", self.api.bind_port)
            self.api.enable_https = http.getboolean("enable_https", self.api.enable_https)

    def _load_security_config(self):
        """Load security.conf"""
        config_file = self.config_dir / "security.conf"
        if not config_file.exists():
            return

        parser = configparser.ConfigParser()
        parser.read(config_file)

        if "Authentication" in parser:
            auth = parser["Authentication"]
            self.security.access_token_expire_minutes = auth.getint(
                "session_timeout",
                self.security.access_token_expire_minutes
            )

        if "Password_Policy" in parser:
            pwd = parser["Password_Policy"]
            self.security.password_min_length = pwd.getint("min_length", self.security.password_min_length)
            self.security.password_expiry_days = pwd.getint("password_expiry_days", self.security.password_expiry_days)

    def get(self, key: str, default: Any = None) -> Any:
        """Get configuration value by key"""
        parts = key.split(".")
        obj = self

        for part in parts:
            if hasattr(obj, part):
                obj = getattr(obj, part)
            else:
                return default

        return obj


# Global config instance
_config: Optional[Config] = None


def get_config(base_dir: Optional[Path] = None) -> Config:
    """Get global configuration instance"""
    global _config
    if _config is None:
        _config = Config(base_dir)
    return _config


def reload_config(base_dir: Optional[Path] = None):
    """Reload configuration"""
    global _config
    _config = Config(base_dir)
    return _config
