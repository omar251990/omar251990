"""
FastAPI Application - Main Entry Point
"""

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import time
import logging

from src.core.config import get_config
from src.core.database import get_engine

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Load configuration
config = get_config()

# Create FastAPI app
app = FastAPI(
    title=config.app.app_name,
    version=config.app.version,
    description="Enterprise Bulk Messaging Platform",
    docs_url="/api/docs",
    redoc_url="/api/redoc"
)

# CORS middleware
if config.api.enable_cors:
    app.add_middleware(
        CORSMiddleware,
        allow_origins=config.api.cors_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )


# Request timing middleware
@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response


# Exception handler
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    logger.error(f"Global exception: {exc}", exc_info=True)
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"}
    )


# Startup event
@app.on_event("startup")
async def startup_event():
    """Run on application startup"""
    logger.info(f"Starting {config.app.app_name} v{config.app.version}")
    logger.info(f"Environment: {config.app.environment}")

    # Initialize database engine
    try:
        engine = get_engine()
        logger.info("Database engine initialized")
    except Exception as e:
        logger.error(f"Failed to initialize database: {e}")


# Shutdown event
@app.on_event("shutdown")
async def shutdown_event():
    """Run on application shutdown"""
    logger.info("Shutting down application...")


# Health check endpoint
@app.get("/api/v1/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "version": config.app.version,
        "build": config.app.build,
        "timestamp": time.time()
    }


# Root endpoint
@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "message": f"Welcome to {config.app.app_name}",
        "version": config.app.version,
        "docs": "/api/docs",
        "health": "/api/v1/health"
    }


# API routes will be added here
# Example:
# from src.api.routes import auth, messages, campaigns
# app.include_router(auth.router, prefix="/api/v1/auth", tags=["auth"])
# app.include_router(messages.router, prefix="/api/v1/messages", tags=["messages"])


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "src.api.main:app",
        host=config.api.bind_address,
        port=config.api.bind_port,
        reload=config.app.environment == "development",
        workers=1 if config.app.environment == "development" else config.app.max_workers
    )
