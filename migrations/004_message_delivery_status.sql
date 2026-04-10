-- Migration: 004_message_delivery_status.sql
-- Description: Add message delivery status tracking (delivered_at field)
-- Created: 2026-04-10
-- ============================================
-- ALTER DIRECT_MESSAGES TABLE - Add delivery tracking
-- ============================================
ALTER TABLE direct_messages
ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP
WITH
    TIME ZONE;

-- Set delivered_at to created_at for existing messages (they were already delivered)
UPDATE direct_messages
SET
    delivered_at = created_at
WHERE
    delivered_at IS NULL;

-- ============================================
-- INDEXES for optimizing status queries
-- ============================================
CREATE INDEX IF NOT EXISTS idx_direct_messages_delivered_at ON direct_messages (delivered_at);

CREATE INDEX IF NOT EXISTS idx_direct_messages_read_at ON direct_messages (read_at);

CREATE INDEX IF NOT EXISTS idx_direct_messages_status ON direct_messages (receiver_id, is_read);