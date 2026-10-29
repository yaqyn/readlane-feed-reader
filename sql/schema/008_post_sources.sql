-- +goose Up
CREATE TABLE post_sources (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, feed_id)
);
CREATE INDEX post_sources_feed_id_idx ON post_sources(feed_id, post_id);
INSERT INTO post_sources (post_id, feed_id)
SELECT id, feed_id FROM posts;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT post_id FROM post_sources GROUP BY post_id HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot roll back post sources while posts belong to multiple feeds';
    END IF;
END $$;
-- +goose StatementEnd
DROP INDEX post_sources_feed_id_idx;
DROP TABLE post_sources;
