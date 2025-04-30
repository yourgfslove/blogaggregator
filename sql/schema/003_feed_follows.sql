-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS feed_follow (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
    UNIQUE(user_id, feed_id)
    );

-- +goose Down
DROP TABLE feed_follow;