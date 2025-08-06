-- Migration: Add QR scan records table
-- SOTA Design: Clean schema following existing patterns

CREATE TABLE IF NOT EXISTS scan_records (
    id VARCHAR(36) PRIMARY KEY,
    courier_id VARCHAR(36) NOT NULL,
    letter_code VARCHAR(20) NOT NULL,
    scan_type VARCHAR(20) NOT NULL CHECK (scan_type IN ('pickup', 'delivery', 'transit')),
    location VARCHAR(255),
    latitude REAL,
    longitude REAL,
    timestamp DATETIME NOT NULL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance - SOTA optimization
CREATE INDEX IF NOT EXISTS idx_scan_records_courier_id ON scan_records(courier_id);
CREATE INDEX IF NOT EXISTS idx_scan_records_letter_code ON scan_records(letter_code);
CREATE INDEX IF NOT EXISTS idx_scan_records_timestamp ON scan_records(timestamp);
CREATE INDEX IF NOT EXISTS idx_scan_records_scan_type ON scan_records(scan_type);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_scan_records_courier_timestamp ON scan_records(courier_id, timestamp DESC);

-- Add columns to existing tables for QR integration
ALTER TABLE letters ADD COLUMN IF NOT EXISTS courier_id VARCHAR(36);
ALTER TABLE letters ADD COLUMN IF NOT EXISTS collected_at DATETIME;
ALTER TABLE letters ADD COLUMN IF NOT EXISTS delivered_at DATETIME;
ALTER TABLE letters ADD COLUMN IF NOT EXISTS delivery_location VARCHAR(255);

-- Add indexes for the new letter columns
CREATE INDEX IF NOT EXISTS idx_letters_courier_id ON letters(courier_id);
CREATE INDEX IF NOT EXISTS idx_letters_collected_at ON letters(collected_at);
CREATE INDEX IF NOT EXISTS idx_letters_delivered_at ON letters(delivered_at);

-- Update courier_tasks table for better QR integration
ALTER TABLE courier_tasks ADD COLUMN IF NOT EXISTS pickup_op_code VARCHAR(6);
ALTER TABLE courier_tasks ADD COLUMN IF NOT EXISTS delivery_op_code VARCHAR(6);
ALTER TABLE courier_tasks ADD COLUMN IF NOT EXISTS current_op_code VARCHAR(6);
ALTER TABLE courier_tasks ADD COLUMN IF NOT EXISTS delivery_notes TEXT;
ALTER TABLE courier_tasks ADD COLUMN IF NOT EXISTS completed_at DATETIME;

-- Add OP Code prefix management to couriers
ALTER TABLE couriers ADD COLUMN IF NOT EXISTS managed_op_code_prefix VARCHAR(6);

-- Create indexes for OP Code queries
CREATE INDEX IF NOT EXISTS idx_courier_tasks_pickup_op ON courier_tasks(pickup_op_code);
CREATE INDEX IF NOT EXISTS idx_courier_tasks_delivery_op ON courier_tasks(delivery_op_code);
CREATE INDEX IF NOT EXISTS idx_couriers_managed_prefix ON couriers(managed_op_code_prefix);