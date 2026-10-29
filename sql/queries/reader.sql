-- name: GetReaderPosts :many
SELECT posts.id, posts.created_at, posts.updated_at, posts.title, posts.url,
       posts.description, posts.published_at, source.feed_id, source.feed_name
FROM posts
JOIN LATERAL (
    SELECT feeds.id AS feed_id, feeds.name AS feed_name
    FROM post_sources
    JOIN feed_follows ON feed_follows.feed_id = post_sources.feed_id
    JOIN feeds ON feeds.id = post_sources.feed_id
    WHERE post_sources.post_id = posts.id
    AND feed_follows.user_id = sqlc.arg(user_id)
    AND (sqlc.narg(source_id)::uuid IS NULL OR post_sources.feed_id = sqlc.narg(source_id)::uuid)
    ORDER BY CASE WHEN feeds.id = posts.feed_id THEN 0 ELSE 1 END, feeds.name, feeds.id
    LIMIT 1
) AS source ON true
WHERE (sqlc.arg(search)::text = '' OR
 to_tsvector('simple', posts.title || ' ' || posts.description) @@ plainto_tsquery('simple', sqlc.arg(search)::text))
ORDER BY posts.published_at DESC NULLS LAST, posts.created_at DESC, posts.id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountReaderPosts :one
SELECT COUNT(*) FROM posts
WHERE EXISTS (
    SELECT 1 FROM post_sources
    JOIN feed_follows ON feed_follows.feed_id = post_sources.feed_id
    WHERE post_sources.post_id = posts.id
    AND feed_follows.user_id = sqlc.arg(user_id)
    AND (sqlc.narg(source_id)::uuid IS NULL OR post_sources.feed_id = sqlc.narg(source_id)::uuid)
)
AND (sqlc.arg(search)::text = '' OR
 to_tsvector('simple', posts.title || ' ' || posts.description) @@ plainto_tsquery('simple', sqlc.arg(search)::text));

-- name: GetReaderPost :one
SELECT posts.id, posts.created_at, posts.updated_at, posts.title, posts.url,
       posts.description, posts.published_at, source.feed_id, source.feed_name
FROM posts
JOIN LATERAL (
    SELECT feeds.id AS feed_id, feeds.name AS feed_name
    FROM post_sources
    JOIN feed_follows ON feed_follows.feed_id = post_sources.feed_id
    JOIN feeds ON feeds.id = post_sources.feed_id
    WHERE post_sources.post_id = posts.id AND feed_follows.user_id = $1
    ORDER BY CASE WHEN feeds.id = posts.feed_id THEN 0 ELSE 1 END, feeds.name, feeds.id
    LIMIT 1
) AS source ON true
WHERE posts.id = $2;

-- name: GetReaderSources :many
SELECT feeds.id, feeds.name FROM feeds
JOIN feed_follows ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1 ORDER BY feeds.name, feeds.id LIMIT 201;
