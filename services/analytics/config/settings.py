import os
from typing import List
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""
    
    # Application
    APP_NAME: str = "UrbanZen Analytics Service"
    DEBUG: bool = False
    LOG_LEVEL: str = "INFO"
    
    # Database
    TIMESCALEDB_URL: str = "postgresql://urbanzen:urbanzen_secure_password@localhost:5433/urbanzen_timeseries"
    MONGODB_URL: str = "mongodb://urbanzen:urbanzen_secure_password@localhost:27017/urbanzen"
    
    # Cache
    REDIS_URL: str = "redis://localhost:6379/0"
    REDIS_PASSWORD: str = "urbanzen_secure_password"
    
    # Message Queue
    KAFKA_BROKERS: List[str] = ["localhost:9092"]
    
    # External APIs
    API_GATEWAY_URL: str = "http://localhost:8080"
    
    # Security
    SECRET_KEY: str = "analytics_secret_key_very_secure"
    ALLOWED_ORIGINS: List[str] = [
        "http://localhost:3000",
        "http://localhost:3001", 
        "http://localhost:3002"
    ]
    
    # ML Models
    MODEL_PATH: str = "/app/models"
    MODEL_CACHE_TTL: int = 3600  # 1 hour
    
    # Analytics
    BATCH_SIZE: int = 1000
    PREDICTION_INTERVAL: int = 300  # 5 minutes
    ANOMALY_THRESHOLD: float = 0.95
    
    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()