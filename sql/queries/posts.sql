-- name: CreatePost :execrows
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (url) DO NOTHING;

-- name: CreatePostSource :exec
INSERT INTO post_sources (post_id, feed_id)
SELECT id, $2 FROM posts WHERE url = $1
ON CONFLICT (post_id, feed_id) DO NOTHING;

-- name: GetPostsForUser :many
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
ORDER BY posts.published_at DESC NULLS LAST, posts.created_at DESC, posts.id
LIMIT $2 OFFSET $3;
