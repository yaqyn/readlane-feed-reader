-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
);
CREATE INDEX posts_feed_published_idx ON posts(feed_id, published_at DESC NULLS LAST, id);

-- +goose Down
DROP TABLE posts;
