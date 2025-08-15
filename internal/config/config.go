package config

import (
    "time"
    "github.com/spf13/viper"
)

type Config struct {
    Environment string `mapstructure:"environment"`
    Version     string `mapstructure:"version"`
    
    Server struct {
        Port         int           `mapstructure:"port"`
        ReadTimeout  time.Duration `mapstructure:"read_timeout"`
        WriteTimeout time.Duration `mapstructure:"write_timeout"`
        IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
    } `mapstructure:"server"`
    
    Database struct {
        Postgres struct {
            Host     string `mapstructure:"host"`
            Port     int    `mapstructure:"port"`
            User     string `mapstructure:"user"`
            Password string `mapstructure:"password"`
            DBName   string `mapstructure:"dbname"`
            SSLMode  string `mapstructure:"sslmode"`
        } `mapstructure:"postgres"`
        
        TimescaleDB struct {
            Host     string `mapstructure:"host"`
            Port     int    `mapstructure:"port"`
            User     string `mapstructure:"user"`
            Password string `mapstructure:"password"`
            DBName   string `mapstructure:"dbname"`
        } `mapstructure:"timescaledb"`
        
        Redis struct {
            Host     string `mapstructure:"host"`
            Port     int    `mapstructure:"port"`
            Password string `mapstructure:"password"`
            DB       int    `mapstructure:"db"`
        } `mapstructure:"redis"`
    } `mapstructure:"database"`
    
    JWT struct {
        Secret    string        `mapstructure:"secret"`
        ExpiresIn time.Duration `mapstructure:"expires_in"`
    } `mapstructure:"jwt"`
    
    Kafka struct {
        Brokers []string `mapstructure:"brokers"`
        Topics  struct {
            DeviceData    string `mapstructure:"device_data"`
            Alerts        string `mapstructure:"alerts"`
            Commands      string `mapstructure:"commands"`
            Notifications string `mapstructure:"notifications"`
        } `mapstructure:"topics"`
    } `mapstructure:"kafka"`
    
    Security struct {
        CORSOrigins      []string `mapstructure:"cors_origins"`
        RateLimitPerMin  int      `mapstructure:"rate_limit_per_min"`
    } `mapstructure:"security"`
    
    Monitoring struct {
        MetricsPort int    `mapstructure:"metrics_port"`
        LogLevel    string `mapstructure:"log_level"`
    } `mapstructure:"monitoring"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")
    
    // Set defaults
    setDefaults()
    
    // Enable environment variable binding
    viper.AutomaticEnv()
    
    // Read config file (optional)
    viper.ReadInConfig()
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

func setDefaults() {
    viper.SetDefault("environment", "development")
    viper.SetDefault("version", "1.0.0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.read_timeout", "30s")
    viper.SetDefault("server.write_timeout", "30s")
    viper.SetDefault("server.idle_timeout", "60s")
    viper.SetDefault("jwt.secret", "default-secret-change-in-production")
    viper.SetDefault("jwt.expires_in", "24h")
    viper.SetDefault("monitoring.metrics_port", 9090)
    viper.SetDefault("monitoring.log_level", "info")
    viper.SetDefault("security.rate_limit_per_min", 100)
    viper.SetDefault("database.postgres.host", "localhost")
    viper.SetDefault("database.postgres.port", 5432)
    viper.SetDefault("database.postgres.user", "postgres")
    viper.SetDefault("database.postgres.password", "password")
    viper.SetDefault("database.postgres.dbname", "urbanzen")
    viper.SetDefault("database.postgres.sslmode", "disable")
    viper.SetDefault("database.redis.host", "localhost")
    viper.SetDefault("database.redis.port", 6379)
    viper.SetDefault("database.redis.db", 0)
    viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
}