-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 200),
    url TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX feeds_user_id_idx ON feeds(user_id);

-- +goose Down
DROP TABLE feeds;
