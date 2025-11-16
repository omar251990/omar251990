"""
Database connection and session management
"""

from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, scoped_session
from sqlalchemy.pool import QueuePool
from contextlib import contextmanager
from typing import Generator

from src.core.config import get_config

# Base class for all models
Base = declarative_base()

# Database engine
_engine = None
_session_factory = None


def get_engine():
    """Get or create database engine"""
    global _engine
    if _engine is None:
        config = get_config()
        _engine = create_engine(
            config.database.url,
            poolclass=QueuePool,
            pool_size=config.database.pool_size,
            max_overflow=20,
            pool_pre_ping=True,
            echo=config.app.environment == "development"
        )
    return _engine


def get_session_factory():
    """Get or create session factory"""
    global _session_factory
    if _session_factory is None:
        engine = get_engine()
        _session_factory = scoped_session(
            sessionmaker(
                autocommit=False,
                autoflush=False,
                bind=engine
            )
        )
    return _session_factory


@contextmanager
def get_db() -> Generator:
    """Get database session with context manager"""
    Session = get_session_factory()
    session = Session()
    try:
        yield session
        session.commit()
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()


def init_db():
    """Initialize database - create all tables"""
    engine = get_engine()
    Base.metadata.create_all(bind=engine)


def close_db():
    """Close database connections"""
    global _engine, _session_factory
    if _session_factory:
        _session_factory.remove()
    if _engine:
        _engine.dispose()
