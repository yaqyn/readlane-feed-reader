-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetFeeds :many
SELECT feeds.*, users.name AS user_name FROM feeds
JOIN users ON users.id = feeds.user_id ORDER BY feeds.name, feeds.id;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds WHERE next_fetch_at <= NOW()
ORDER BY last_fetched_at ASC NULLS FIRST, id LIMIT 1;

-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = NOW(), updated_at = NOW(), next_fetch_at = NOW() + INTERVAL '1 minute' WHERE id = $1;

-- name: ScheduleFeedSuccess :exec
UPDATE feeds SET next_fetch_at = $2, last_success_at = NOW(), failure_count = 0,
last_error = '', etag = $3, last_modified = $4 WHERE id = $1;

-- name: ScheduleFeedFailure :exec
UPDATE feeds SET next_fetch_at = $2, failure_count = LEAST(failure_count + 1, 30),
last_error = $3 WHERE id = $1;
