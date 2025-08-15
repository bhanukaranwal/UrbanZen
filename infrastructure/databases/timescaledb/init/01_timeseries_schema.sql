-- TimescaleDB Schema for UrbanZen Smart City Platform
-- Time-series database for sensor data and telemetry

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Device telemetry data (main time-series table)
CREATE TABLE device_telemetry (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(20),
    quality_score FLOAT DEFAULT 1.0, -- Data quality score (0-1)
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Convert to hypertable (partitioned by time)
SELECT create_hypertable('device_telemetry', 'time', chunk_time_interval => INTERVAL '1 day');

-- Create indexes for efficient querying
CREATE INDEX idx_device_telemetry_device_id_time ON device_telemetry (device_id, time DESC);
CREATE INDEX idx_device_telemetry_metric_time ON device_telemetry (metric_name, time DESC);
CREATE INDEX idx_device_telemetry_device_metric ON device_telemetry (device_id, metric_name, time DESC);

-- Water utility metrics
CREATE TABLE water_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    flow_rate DOUBLE PRECISION, -- L/min
    pressure DOUBLE PRECISION, -- bar
    temperature DOUBLE PRECISION, -- Celsius
    ph_level DOUBLE PRECISION,
    turbidity DOUBLE PRECISION, -- NTU
    chlorine_level DOUBLE PRECISION, -- mg/L
    total_dissolved_solids DOUBLE PRECISION, -- ppm
    leak_detected BOOLEAN DEFAULT FALSE,
    valve_position DOUBLE PRECISION, -- percentage (0-100)
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('water_metrics', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_water_metrics_device_time ON water_metrics (device_id, time DESC);

-- Electricity utility metrics
CREATE TABLE electricity_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    voltage_l1 DOUBLE PRECISION, -- Volts
    voltage_l2 DOUBLE PRECISION,
    voltage_l3 DOUBLE PRECISION,
    current_l1 DOUBLE PRECISION, -- Amperes
    current_l2 DOUBLE PRECISION,
    current_l3 DOUBLE PRECISION,
    power_active DOUBLE PRECISION, -- kW
    power_reactive DOUBLE PRECISION, -- kVAR
    power_apparent DOUBLE PRECISION, -- kVA
    power_factor DOUBLE PRECISION,
    frequency DOUBLE PRECISION, -- Hz
    energy_consumed DOUBLE PRECISION, -- kWh (cumulative)
    thd_voltage DOUBLE PRECISION, -- Total Harmonic Distortion
    thd_current DOUBLE PRECISION,
    temperature DOUBLE PRECISION, -- Celsius
    fault_detected BOOLEAN DEFAULT FALSE,
    fault_code VARCHAR(20),
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('electricity_metrics', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_electricity_metrics_device_time ON electricity_metrics (device_id, time DESC);

-- Traffic and transport metrics
CREATE TABLE traffic_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    vehicle_count INTEGER DEFAULT 0,
    average_speed DOUBLE PRECISION, -- km/h
    congestion_level VARCHAR(20), -- low, medium, high, critical
    vehicle_types JSONB DEFAULT '{}'::jsonb, -- {"cars": 10, "bikes": 5, "trucks": 2}
    incident_detected BOOLEAN DEFAULT FALSE,
    incident_type VARCHAR(50),
    weather_condition VARCHAR(50),
    visibility DOUBLE PRECISION, -- meters
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('traffic_metrics', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_traffic_metrics_device_time ON traffic_metrics (device_id, time DESC);

-- Environmental metrics
CREATE TABLE environmental_metrics (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    temperature DOUBLE PRECISION, -- Celsius
    humidity DOUBLE PRECISION, -- percentage
    pressure DOUBLE PRECISION, -- hPa
    pm25 DOUBLE PRECISION, -- μg/m³
    pm10 DOUBLE PRECISION, -- μg/m³
    co2 DOUBLE PRECISION, -- ppm
    co DOUBLE PRECISION, -- ppm
    no2 DOUBLE PRECISION, -- ppb
    so2 DOUBLE PRECISION, -- ppb
    o3 DOUBLE PRECISION, -- ppb
    noise_level DOUBLE PRECISION, -- dB
    uv_index DOUBLE PRECISION,
    wind_speed DOUBLE PRECISION, -- m/s
    wind_direction DOUBLE PRECISION, -- degrees
    rainfall DOUBLE PRECISION, -- mm
    air_quality_index INTEGER,
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('environmental_metrics', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_environmental_metrics_device_time ON environmental_metrics (device_id, time DESC);

-- Energy consumption aggregates
CREATE TABLE energy_consumption_hourly (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    consumption_kwh DOUBLE PRECISION NOT NULL,
    peak_demand_kw DOUBLE PRECISION,
    cost DOUBLE PRECISION,
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('energy_consumption_hourly', 'time', chunk_time_interval => INTERVAL '7 days');

-- Water consumption aggregates
CREATE TABLE water_consumption_hourly (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    consumption_liters DOUBLE PRECISION NOT NULL,
    peak_flow_rate DOUBLE PRECISION,
    cost DOUBLE PRECISION,
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('water_consumption_hourly', 'time', chunk_time_interval => INTERVAL '7 days');

-- System alerts and events
CREATE TABLE system_alerts (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    alert_type VARCHAR(50) NOT NULL, -- anomaly, fault, maintenance, security
    severity VARCHAR(20) NOT NULL, -- low, medium, high, critical
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'open', -- open, acknowledged, resolved
    acknowledged_by VARCHAR(100),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('system_alerts', 'time', chunk_time_interval => INTERVAL '7 days');
CREATE INDEX idx_system_alerts_device_time ON system_alerts (device_id, time DESC);
CREATE INDEX idx_system_alerts_type_severity ON system_alerts (alert_type, severity, time DESC);

-- Device commands and responses
CREATE TABLE device_commands (
    time TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    command_id VARCHAR(100) UNIQUE NOT NULL,
    command_type VARCHAR(50) NOT NULL,
    command_data JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, acknowledged, executed, failed
    response_data JSONB,
    sent_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

SELECT create_hypertable('device_commands', 'time', chunk_time_interval => INTERVAL '7 days');
CREATE INDEX idx_device_commands_device_time ON device_commands (device_id, time DESC);
CREATE INDEX idx_device_commands_id ON device_commands (command_id);

-- Create continuous aggregates for analytics
-- Hourly averages for device telemetry
CREATE MATERIALIZED VIEW device_telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS bucket,
       device_id,
       metric_name,
       AVG(metric_value) AS avg_value,
       MIN(metric_value) AS min_value,
       MAX(metric_value) AS max_value,
       COUNT(*) AS sample_count
FROM device_telemetry
GROUP BY bucket, device_id, metric_name;

-- Daily aggregates for consumption data
CREATE MATERIALIZED VIEW water_consumption_daily
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 day', time) AS bucket,
       device_id,
       SUM(consumption_liters) AS total_consumption,
       AVG(peak_flow_rate) AS avg_peak_flow,
       SUM(cost) AS total_cost
FROM water_consumption_hourly
GROUP BY bucket, device_id;

CREATE MATERIALIZED VIEW energy_consumption_daily
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 day', time) AS bucket,
       device_id,
       SUM(consumption_kwh) AS total_consumption,
       AVG(peak_demand_kw) AS avg_peak_demand,
       SUM(cost) AS total_cost
FROM energy_consumption_hourly
GROUP BY bucket, device_id;

-- Environmental quality hourly aggregates
CREATE MATERIALIZED VIEW environmental_quality_hourly
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS bucket,
       device_id,
       AVG(temperature) AS avg_temperature,
       AVG(humidity) AS avg_humidity,
       AVG(pm25) AS avg_pm25,
       AVG(pm10) AS avg_pm10,
       AVG(co2) AS avg_co2,
       AVG(noise_level) AS avg_noise_level,
       AVG(air_quality_index) AS avg_aqi
FROM environmental_metrics
GROUP BY bucket, device_id;

-- Set up retention policies (automatically drop old data)
-- Keep raw telemetry data for 1 year
SELECT add_retention_policy('device_telemetry', INTERVAL '1 year');

-- Keep raw utility metrics for 2 years
SELECT add_retention_policy('water_metrics', INTERVAL '2 years');
SELECT add_retention_policy('electricity_metrics', INTERVAL '2 years');

-- Keep environmental and traffic data for 1 year
SELECT add_retention_policy('environmental_metrics', INTERVAL '1 year');
SELECT add_retention_policy('traffic_metrics', INTERVAL '1 year');

-- Keep alerts for 3 years (compliance requirement)
SELECT add_retention_policy('system_alerts', INTERVAL '3 years');

-- Keep commands for 6 months
SELECT add_retention_policy('device_commands', INTERVAL '6 months');

-- Create indexes for common query patterns
CREATE INDEX idx_device_telemetry_recent ON device_telemetry (time DESC, device_id) WHERE time > NOW() - INTERVAL '7 days';
CREATE INDEX idx_water_metrics_recent ON water_metrics (time DESC, device_id) WHERE time > NOW() - INTERVAL '7 days';
CREATE INDEX idx_electricity_metrics_recent ON electricity_metrics (time DESC, device_id) WHERE time > NOW() - INTERVAL '7 days';