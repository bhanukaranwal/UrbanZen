package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	LogLevel    string
	
	Database struct {
		PostgresURL     string
		TimescaleDBURL  string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}
	
	MQTT struct {
		Broker   string
		Username string
		Password string
		ClientID string
	}
	
	API struct {
		GatewayURL string
		APIKey     string
	}
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8081"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	// Database configuration
	cfg.Database.PostgresURL = getEnv("POSTGRES_URL", "postgres://urbanzen:urbanzen_secure_password@localhost:5432/urbanzen?sslmode=disable")
	cfg.Database.TimescaleDBURL = getEnv("TIMESCALEDB_URL", "postgres://urbanzen:urbanzen_secure_password@localhost:5433/urbanzen_timeseries?sslmode=disable")
	cfg.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	cfg.Database.ConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)

	// MQTT configuration
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "device-mgmt-service")

	// API configuration
	cfg.API.GatewayURL = getEnv("API_GATEWAY_URL", "http://localhost:8080")
	cfg.API.APIKey = getEnv("API_KEY", "device_mgmt_api_key")

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}