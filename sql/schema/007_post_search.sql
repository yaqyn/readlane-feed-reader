-- +goose Up
CREATE INDEX posts_search_idx ON posts USING GIN
(to_tsvector('simple', title || ' ' || description));

-- +goose Down
DROP INDEX posts_search_idx;
