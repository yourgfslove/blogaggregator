-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    published_at TIMESTAMP,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description  TEXT,
    feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE
    );

-- +goose Down
DROP TABLE posts;