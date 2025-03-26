-- File: 001_init_schema.up.sql
-- This file creates the initial database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Devices table to store device information
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id VARCHAR(255) UNIQUE NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Health records table to store health measurements
CREATE TABLE health_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id VARCHAR(255) NOT NULL,
    spo2 INTEGER CHECK (spo2 BETWEEN 0 AND 100),
    pulse INTEGER CHECK (pulse BETWEEN 0 AND 300),
    temperature DECIMAL(4,1) CHECK (temperature BETWEEN 95.0 AND 106.0),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_health_records_device_id ON health_records(device_id);
CREATE INDEX idx_health_records_recorded_at ON health_records(recorded_at DESC);

-- Optional: Add a function to automatically update last_active timestamp
CREATE OR REPLACE FUNCTION update_last_active()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE devices 
    SET last_active = CURRENT_TIMESTAMP 
    WHERE device_id = NEW.device_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update last_active when a new health record is inserted
CREATE TRIGGER update_device_last_active
AFTER INSERT ON health_records
FOR EACH ROW
EXECUTE FUNCTION update_last_active();