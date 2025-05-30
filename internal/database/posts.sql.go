// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: posts.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createPost = `-- name: CreatePost :exec
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
)
`

type CreatePostParams struct {
	ID          uuid.UUID
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	PublishedAt sql.NullTime
	Title       string
	Url         string
	Description sql.NullString
	FeedID      uuid.UUID
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) error {
	_, err := q.db.ExecContext(ctx, createPost,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.PublishedAt,
		arg.Title,
		arg.Url,
		arg.Description,
		arg.FeedID,
	)
	return err
}

const getPosts = `-- name: GetPosts :many
SELECT posts.id, posts.created_at, posts.updated_at, published_at, title, url, description, posts.feed_id, feed_follow.id, feed_follow.created_at, feed_follow.updated_at, user_id, feed_follow.feed_id
FROM posts
JOIN feed_follow ON posts.feed_id = feed_follow.feed_id
WHERE feed_follow.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2
`

type GetPostsParams struct {
	UserID uuid.UUID
	Limit  int32
}

type GetPostsRow struct {
	ID          uuid.UUID
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	PublishedAt sql.NullTime
	Title       string
	Url         string
	Description sql.NullString
	FeedID      uuid.UUID
	ID_2        uuid.UUID
	CreatedAt_2 sql.NullTime
	UpdatedAt_2 sql.NullTime
	UserID      uuid.UUID
	FeedID_2    uuid.UUID
}

func (q *Queries) GetPosts(ctx context.Context, arg GetPostsParams) ([]GetPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPosts, arg.UserID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsRow
	for rows.Next() {
		var i GetPostsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.PublishedAt,
			&i.Title,
			&i.Url,
			&i.Description,
			&i.FeedID,
			&i.ID_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.UserID,
			&i.FeedID_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
