-- name: GetReaderPosts :many
SELECT posts.*, feeds.name AS feed_name FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
JOIN feeds ON feeds.id = posts.feed_id
WHERE feed_follows.user_id = sqlc.arg(user_id)
AND (sqlc.narg(source_id)::uuid IS NULL OR posts.feed_id = sqlc.narg(source_id)::uuid)
AND (sqlc.arg(search)::text = '' OR
 to_tsvector('simple', posts.title || ' ' || posts.description) @@ plainto_tsquery('simple', sqlc.arg(search)::text))
ORDER BY posts.published_at DESC NULLS LAST, posts.created_at DESC, posts.id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountReaderPosts :one
SELECT COUNT(*) FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
WHERE feed_follows.user_id = sqlc.arg(user_id)
AND (sqlc.narg(source_id)::uuid IS NULL OR posts.feed_id = sqlc.narg(source_id)::uuid)
AND (sqlc.arg(search)::text = '' OR
 to_tsvector('simple', posts.title || ' ' || posts.description) @@ plainto_tsquery('simple', sqlc.arg(search)::text));

-- name: GetReaderPost :one
SELECT posts.*, feeds.name AS feed_name FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
JOIN feeds ON feeds.id = posts.feed_id
WHERE feed_follows.user_id = $1 AND posts.id = $2;

-- name: GetReaderSources :many
SELECT feeds.id, feeds.name FROM feeds
JOIN feed_follows ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1 ORDER BY feeds.name, feeds.id LIMIT 201;
