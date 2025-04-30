-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;