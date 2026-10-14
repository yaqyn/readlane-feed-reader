-- +goose Up
ALTER TABLE feeds
    ADD COLUMN next_fetch_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN last_success_at TIMESTAMPTZ,
    ADD COLUMN failure_count INTEGER NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    ADD COLUMN last_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN etag TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_modified TEXT NOT NULL DEFAULT '';
CREATE INDEX feeds_due_idx ON feeds(next_fetch_at);

-- +goose Down
DROP INDEX feeds_due_idx;
ALTER TABLE feeds
    DROP COLUMN last_modified,
    DROP COLUMN etag,
    DROP COLUMN last_error,
    DROP COLUMN failure_count,
    DROP COLUMN last_success_at,
    DROP COLUMN next_fetch_at;
