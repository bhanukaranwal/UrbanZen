#!/usr/bin/env python3
"""
UrbanZen Analytics Service
AI-powered analytics and insights for IoT smart city data
"""

import os
import sys
import asyncio
import logging
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from prometheus_client import make_asgi_app

# Add the app directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), 'app'))

from app.api.routes import api_router
from app.services.database import database_manager
from app.services.cache import cache_manager
from app.ml.model_manager import model_manager
from config.settings import settings


# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL.upper()),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan events"""
    # Startup
    logger.info("Starting UrbanZen Analytics Service")
    
    # Initialize database connections
    await database_manager.connect()
    
    # Initialize cache
    await cache_manager.connect()
    
    # Load ML models
    await model_manager.load_models()
    
    logger.info("Analytics service started successfully")
    
    yield
    
    # Shutdown
    logger.info("Shutting down Analytics service")
    
    # Close database connections
    await database_manager.disconnect()
    
    # Close cache connection
    await cache_manager.disconnect()
    
    logger.info("Analytics service shutdown complete")


# Create FastAPI application
app = FastAPI(
    title="UrbanZen Analytics Service",
    description="AI-powered analytics and insights for IoT smart city data",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan
)

# Add middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.add_middleware(GZipMiddleware, minimum_size=1000)

# Include API routes
app.include_router(api_router, prefix="/api/v1")

# Add Prometheus metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "analytics",
        "version": "1.0.0",
        "models_loaded": await model_manager.get_model_status()
    }


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "UrbanZen Analytics Service",
        "version": "1.0.0",
        "description": "AI-powered analytics and insights for IoT smart city data",
        "docs": "/docs",
        "health": "/health"
    }


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=int(os.getenv("PORT", "8083")),
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower(),
        access_log=True
    )