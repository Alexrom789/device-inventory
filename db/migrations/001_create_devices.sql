-- Migration: Create devices table
-- Run this manually in psql before starting the server

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS devices (
    id          VARCHAR(36)     PRIMARY KEY,
    imei        VARCHAR(20)     NOT NULL UNIQUE,
    model       VARCHAR(100)    NOT NULL,
    status      VARCHAR(20)     NOT NULL DEFAULT 'received',
    grade       VARCHAR(10)     NOT NULL DEFAULT 'ungraded',
    price       DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    created_at  TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Index on status to speed up filtered queries like "show me all received devices"
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);

-- Index on IMEI since it's a common lookup field
CREATE INDEX IF NOT EXISTS idx_devices_imei ON devices(imei);
