-- PostgreSQL Database Schema for UrbanZen Smart City Platform
-- Main relational database for user management, device registry, and system configuration

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Users table with RBAC support
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL DEFAULT 'citizen',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Create indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Device types and categories
CREATE TABLE device_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL, -- water, electricity, transport, environment
    description TEXT,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    version VARCHAR(50),
    capabilities JSONB DEFAULT '{}'::jsonb,
    configuration_schema JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Device registry
CREATE TABLE devices (
    id BIGSERIAL PRIMARY KEY,
    device_id VARCHAR(100) UNIQUE NOT NULL,
    device_type_id BIGINT REFERENCES device_types(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    location GEOMETRY(POINT, 4326), -- GPS coordinates
    address TEXT,
    ward_id INTEGER,
    zone_id INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive', -- active, inactive, maintenance, error
    connectivity_status VARCHAR(20) DEFAULT 'disconnected', -- connected, disconnected, unknown
    configuration JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    installed_at TIMESTAMP WITH TIME ZONE,
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create spatial index for devices
CREATE INDEX idx_devices_location ON devices USING GIST(location);
CREATE INDEX idx_devices_device_id ON devices(device_id);
CREATE INDEX idx_devices_type ON devices(device_type_id);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_connectivity ON devices(connectivity_status);

-- API keys for service-to-service communication
CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '{}'::jsonb,
    active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    last_used TIMESTAMP WITH TIME ZONE
);

-- Insert default system data
INSERT INTO device_types (name, category, description, manufacturer, capabilities) VALUES
('Smart Water Meter', 'water', 'IoT-enabled water consumption meter', 'AquaTech Solutions', '{"measurement": "flow_rate", "units": "liters", "frequency": "real_time"}'),
('Smart Electricity Meter', 'electricity', 'IoT-enabled electricity consumption meter', 'PowerSense Systems', '{"measurement": "power", "units": "kWh", "frequency": "real_time", "power_quality": true}'),
('Traffic Camera', 'transport', 'AI-enabled traffic monitoring camera', 'VisionTech', '{"video_resolution": "4K", "ai_analytics": true, "night_vision": true}'),
('Air Quality Sensor', 'environment', 'Multi-parameter air quality monitoring sensor', 'EnviroSense', '{"parameters": ["PM2.5", "PM10", "CO2", "NO2", "SO2"], "accuracy": "±5%"}'),
('Smart Streetlight', 'infrastructure', 'IoT-enabled LED streetlight with sensors', 'LightSmart', '{"dimming": true, "motion_sensor": true, "energy_monitoring": true}'),
('Waste Level Sensor', 'sanitation', 'Ultrasonic waste bin level monitoring sensor', 'CleanTech', '{"measurement": "fill_level", "accuracy": "±2cm", "battery_life": "5_years"}')