-- +goose Up
ALTER TABLE feeds ADD COLUMN last_fetched_at TIMESTAMPTZ;
CREATE INDEX feeds_fetch_order_idx ON feeds(last_fetched_at ASC NULLS FIRST, id);

-- +goose Down
DROP INDEX feeds_fetch_order_idx;
ALTER TABLE feeds DROP COLUMN last_fetched_at;
