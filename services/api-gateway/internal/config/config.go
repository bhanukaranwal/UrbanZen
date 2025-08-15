package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Environment string
	Port        string
	LogLevel    string
	
	Database struct {
		URL            string
		MaxOpenConns   int
		MaxIdleConns   int
		ConnMaxLifetime int
	}
	
	Redis struct {
		URL      string
		Password string
		DB       int
	}
	
	JWT struct {
		Secret        string
		ExpireMinutes int
	}
	
	Server struct {
		ReadTimeout  int
		WriteTimeout int
		IdleTimeout  int
	}
	
	RateLimit struct {
		RequestsPerMinute int
		BurstSize         int
	}
	
	Services struct {
		DeviceManagement string
		DataIngestion    string
		Analytics        string
		Notification     string
		UserManagement   string
		Billing          string
		Reporting        string
	}
	
	Security struct {
		AllowedOrigins []string
		TLSCertFile    string
		TLSKeyFile     string
	}
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	// Database configuration
	cfg.Database.URL = getEnv("POSTGRES_URL", "postgres://urbanzen:urbanzen_secure_password@localhost:5432/urbanzen?sslmode=disable")
	cfg.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	cfg.Database.ConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)

	// Redis configuration
	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379/0")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "urbanzen_secure_password")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)

	// JWT configuration
	cfg.JWT.Secret = getEnv("JWT_SECRET", "urbanzen_jwt_secret_key_very_secure")
	cfg.JWT.ExpireMinutes = getEnvAsInt("JWT_EXPIRE_MINUTES", 60)

	// Server configuration
	cfg.Server.ReadTimeout = getEnvAsInt("SERVER_READ_TIMEOUT", 30)
	cfg.Server.WriteTimeout = getEnvAsInt("SERVER_WRITE_TIMEOUT", 30)
	cfg.Server.IdleTimeout = getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)

	// Rate limiting configuration
	cfg.RateLimit.RequestsPerMinute = getEnvAsInt("RATE_LIMIT_RPM", 100)
	cfg.RateLimit.BurstSize = getEnvAsInt("RATE_LIMIT_BURST", 50)

	// Microservices URLs
	cfg.Services.DeviceManagement = getEnv("DEVICE_MGMT_URL", "http://localhost:8081")
	cfg.Services.DataIngestion = getEnv("DATA_INGESTION_URL", "http://localhost:8082")
	cfg.Services.Analytics = getEnv("ANALYTICS_URL", "http://localhost:8083")
	cfg.Services.Notification = getEnv("NOTIFICATION_URL", "http://localhost:8084")
	cfg.Services.UserManagement = getEnv("USER_MGMT_URL", "http://localhost:8085")
	cfg.Services.Billing = getEnv("BILLING_URL", "http://localhost:8086")
	cfg.Services.Reporting = getEnv("REPORTING_URL", "http://localhost:8087")

	// Security configuration
	cfg.Security.AllowedOrigins = []string{
		getEnv("ALLOWED_ORIGIN_1", "http://localhost:3000"),
		getEnv("ALLOWED_ORIGIN_2", "http://localhost:3001"),
		getEnv("ALLOWED_ORIGIN_3", "http://localhost:3002"),
	}
	cfg.Security.TLSCertFile = getEnv("TLS_CERT_FILE", "")
	cfg.Security.TLSKeyFile = getEnv("TLS_KEY_FILE", "")

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}