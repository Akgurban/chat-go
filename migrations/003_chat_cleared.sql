-- Track when users cleared their DM chats
CREATE TABLE IF NOT EXISTS chat_cleared (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    other_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cleared_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, other_user_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_cleared_user ON chat_cleared(user_id);
