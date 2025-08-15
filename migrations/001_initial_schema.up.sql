-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'citizen',
    phone VARCHAR(20),
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    notification_preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Devices table
CREATE TABLE devices (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    status VARCHAR(50) DEFAULT 'active',
    firmware_version VARCHAR(50),
    hardware_version VARCHAR(50),
    installation_date DATE,
    last_maintenance DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Alerts table
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    device_id VARCHAR(255),
    user_id UUID,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by UUID,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (device_id) REFERENCES devices(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (acknowledged_by) REFERENCES users(id)
);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    priority VARCHAR(50) NOT NULL DEFAULT 'normal',
    channels JSONB DEFAULT '[]',
    scheduled_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create indexes
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_location ON devices USING GIST(location);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_acknowledged ON alerts(acknowledged);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);

-- Create audit log function
CREATE OR REPLACE FUNCTION audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add audit triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER update_devices_updated_at
    BEFORE UPDATE ON devices
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger();

-- Insert sample data
INSERT INTO users (username, email, password_hash, first_name, last_name, role) VALUES
('admin', 'admin@urbanzen.gov.in', '$2a$10$8K1p/a0dEVV0kHiP7KaGOuZl4/DlHc2RYH6.X9qf6H.xP1aZ8/L3G', 'Admin', 'User', 'admin'),
('operator', 'operator@urbanzen.gov.in', '$2a$10$8K1p/a0dEVV0kHiP7KaGOuZl4/DlHc2RYH6.X9qf6H.xP1aZ8/L3G', 'System', 'Operator', 'operator'),
('citizen1', 'citizen1@example.com', '$2a$10$8K1p/a0dEVV0kHiP7KaGOuZl4/DlHc2RYH6.X9qf6H.xP1aZ8/L3G', 'John', 'Doe', 'citizen');

INSERT INTO devices (id, name, type, location) VALUES
('WS-001', 'Water Sensor Connaught Place', 'water_sensor', ST_GeogFromText('POINT(77.2167 28.6328)')),
('EM-001', 'Smart Meter India Gate', 'electricity_meter', ST_GeogFromText('POINT(77.2295 28.6129)')),
('TC-001', 'Traffic Camera Red Fort', 'traffic_camera', ST_GeogFromText('POINT(77.2410 28.6562)'));