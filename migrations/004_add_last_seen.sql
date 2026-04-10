-- Migration: 004_add_last_seen.sql
-- Description: Add last_seen_at column to track when offline users were last active
-- Created: 2026-04-09

ALTER TABLE users ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP WITH TIME ZONE;

-- Set initial last_seen_at for existing offline users to their last updated_at
UPDATE users SET last_seen_at = updated_at WHERE status = 'offline' AND last_seen_at IS NULL;
