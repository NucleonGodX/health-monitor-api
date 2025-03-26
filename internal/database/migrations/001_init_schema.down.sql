-- File: 001_init_schema.down.sql
-- This file removes the database schema (used for rollbacks)

-- Drop trigger first
DROP TRIGGER IF EXISTS update_device_last_active ON health_records;

-- Drop function
DROP FUNCTION IF EXISTS update_last_active();

-- Drop tables
DROP TABLE IF EXISTS health_records;
DROP TABLE IF EXISTS devices;

-- Remove UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";