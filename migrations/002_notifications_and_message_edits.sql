-- Migration: 002_notifications_and_message_edits.sql
-- Description: Add notifications, push subscriptions, message read status, and edit support
-- Created: 2026-04-04

-- ============================================
-- NOTIFICATIONS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data TEXT,
    is_read BOOLEAN DEFAULT false,
    is_pushed BOOLEAN DEFAULT false,
    reference_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP WITH TIME ZONE
);

-- ============================================
-- PUSH SUBSCRIPTIONS TABLE (for Web Push notifications)
-- ============================================
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    p256dh VARCHAR(255) NOT NULL,
    auth VARCHAR(255) NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, endpoint)
);

-- ============================================
-- NOTIFICATION PREFERENCES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    direct_message_notify BOOLEAN DEFAULT true,
    mention_notify BOOLEAN DEFAULT true,
    room_message_notify BOOLEAN DEFAULT true,
    mute_all BOOLEAN DEFAULT false,
    quiet_hours_enabled BOOLEAN DEFAULT false,
    quiet_hours_start VARCHAR(5),
    quiet_hours_end VARCHAR(5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- MESSAGE READ STATUS TABLE (for room messages)
-- ============================================
CREATE TABLE IF NOT EXISTS message_read_status (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, message_id)
);

-- ============================================
-- ALTER MESSAGES TABLE - Add edit support
-- ============================================
ALTER TABLE messages 
    ADD COLUMN IF NOT EXISTS is_edited BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS edited_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;

-- ============================================
-- ALTER DIRECT_MESSAGES TABLE - Add edit support and read_at
-- ============================================
ALTER TABLE direct_messages 
    ADD COLUMN IF NOT EXISTS is_edited BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS edited_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS read_at TIMESTAMP WITH TIME ZONE;

-- ============================================
-- INDEXES
-- ============================================
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_id ON push_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_message_read_status_user_id ON message_read_status(user_id);
CREATE INDEX IF NOT EXISTS idx_message_read_status_message_id ON message_read_status(message_id);
CREATE INDEX IF NOT EXISTS idx_message_read_status_room_id ON message_read_status(room_id, user_id);
CREATE INDEX IF NOT EXISTS idx_messages_is_deleted ON messages(is_deleted);
CREATE INDEX IF NOT EXISTS idx_direct_messages_is_deleted ON direct_messages(is_deleted);
