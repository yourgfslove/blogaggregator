-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, published_at, title, url, description, feed_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
);

-- name: GetPosts :many
SELECT *
FROM posts
JOIN feed_follow ON posts.feed_id = feed_follow.feed_id
WHERE feed_follow.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;