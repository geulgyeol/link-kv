-- =============
-- blog_users
-- =============

-- name: GetBlogUser :one
SELECT * FROM blog_users WHERE blog_platform = $1 AND user_id = $2;

-- name: CreateBlogUser :one
INSERT INTO blog_users (blog_platform, user_id) VALUES ($1, $2)
ON CONFLICT (blog_platform, user_id) DO NOTHING
RETURNING *;

-- name: DeleteBlogUser :exec
DELETE FROM blog_users WHERE blog_platform = $1 AND user_id = $2;

-- name: UpdateBlogUserLastEnqueuedAt :exec
UPDATE blog_users SET last_enqueued_at = current_timestamp
WHERE blog_platform = $1 AND user_id = $2;

-- name: ListBlogUsers :many
SELECT * FROM blog_users
ORDER BY created_at DESC
LIMIT $1;

-- name: ListBlogUsersOlderThan :many
SELECT * FROM blog_users
WHERE last_enqueued_at < $1
ORDER BY last_enqueued_at ASC
LIMIT $2;

-- =============
-- blog_posts
-- =============

-- name: GetBlogPost :one
SELECT * FROM blog_posts WHERE post_url = $1;

-- name: ListBlogPosts :many
SELECT * FROM blog_posts
ORDER BY created_at DESC
LIMIT $1;

-- name: ListBlogPostsByPlatform :many
SELECT * FROM blog_posts
WHERE blog_platform = $1
ORDER BY created_at DESC
LIMIT $2;
