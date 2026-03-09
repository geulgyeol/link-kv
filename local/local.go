package local

import (
	"context"
	"sync"
	"time"

	"github.com/geulgyeol/link-kv/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// LinkKVService defines the interface for link-kv operations
type LinkKVService interface {
	// Blog Users
	GetBlogUser(ctx context.Context, arg db.GetBlogUserParams) (db.BlogUser, error)
	CreateBlogUser(ctx context.Context, arg db.CreateBlogUserParams) (db.BlogUser, error)
	DeleteBlogUser(ctx context.Context, arg db.DeleteBlogUserParams) error
	UpdateBlogUserLastEnqueuedAt(ctx context.Context, arg db.UpdateBlogUserLastEnqueuedAtParams) error
	ListBlogUsers(ctx context.Context, limit int32) ([]db.BlogUser, error)

	// Blog Posts
	GetBlogPost(ctx context.Context, postUrl string) (db.BlogPost, error)
	ListBlogPosts(ctx context.Context, limit int32) ([]db.BlogPost, error)
	ListBlogPostsByPlatform(ctx context.Context, arg db.ListBlogPostsByPlatformParams) ([]db.BlogPost, error)
}

// LocalQueries implements in-memory storage for local development/testing
type LocalQueries struct {
	mu        sync.Mutex
	blogUsers []db.BlogUser
	blogPosts []db.BlogPost
}

func New() *LocalQueries {
	return &LocalQueries{
		blogUsers: make([]db.BlogUser, 0),
		blogPosts: make([]db.BlogPost, 0),
	}
}

// Blog User Operations

func (q *LocalQueries) GetBlogUser(_ context.Context, arg db.GetBlogUserParams) (db.BlogUser, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, user := range q.blogUsers {
		if user.BlogPlatform == arg.BlogPlatform && user.UserID == arg.UserID {
			return user, nil
		}
	}

	return db.BlogUser{}, pgx.ErrNoRows
}

func (q *LocalQueries) CreateBlogUser(_ context.Context, arg db.CreateBlogUserParams) (db.BlogUser, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// ON CONFLICT DO NOTHING — if already exists, return no rows (same as pgx behavior)
	for _, user := range q.blogUsers {
		if user.BlogPlatform == arg.BlogPlatform && user.UserID == arg.UserID {
			return db.BlogUser{}, pgx.ErrNoRows
		}
	}

	now := time.Now()
	user := db.BlogUser{
		BlogPlatform:   arg.BlogPlatform,
		UserID:         arg.UserID,
		LastEnqueuedAt: pgtype.Timestamp{Time: now, Valid: true},
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
	}
	q.blogUsers = append(q.blogUsers, user)

	return user, nil
}

func (q *LocalQueries) DeleteBlogUser(_ context.Context, arg db.DeleteBlogUserParams) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	newUsers := make([]db.BlogUser, 0, len(q.blogUsers))
	for _, user := range q.blogUsers {
		if user.BlogPlatform == arg.BlogPlatform && user.UserID == arg.UserID {
			continue
		}
		newUsers = append(newUsers, user)
	}
	q.blogUsers = newUsers

	return nil
}

func (q *LocalQueries) UpdateBlogUserLastEnqueuedAt(_ context.Context, arg db.UpdateBlogUserLastEnqueuedAtParams) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := range q.blogUsers {
		if q.blogUsers[i].BlogPlatform == arg.BlogPlatform && q.blogUsers[i].UserID == arg.UserID {
			q.blogUsers[i].LastEnqueuedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
			return nil
		}
	}

	return nil
}

func (q *LocalQueries) ListBlogUsers(_ context.Context, limit int32) ([]db.BlogUser, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Return in reverse order (newest first, like ORDER BY created_at DESC)
	count := len(q.blogUsers)
	if int(limit) < count {
		count = int(limit)
	}

	result := make([]db.BlogUser, count)
	for i := 0; i < count; i++ {
		result[i] = q.blogUsers[len(q.blogUsers)-1-i]
	}

	return result, nil
}

// Blog Post Operations

func (q *LocalQueries) GetBlogPost(_ context.Context, postUrl string) (db.BlogPost, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, post := range q.blogPosts {
		if post.PostUrl == postUrl {
			return post, nil
		}
	}

	return db.BlogPost{}, pgx.ErrNoRows
}

func (q *LocalQueries) ListBlogPosts(_ context.Context, limit int32) ([]db.BlogPost, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Return in reverse order (newest first, like ORDER BY created_at DESC)
	count := len(q.blogPosts)
	if int(limit) < count {
		count = int(limit)
	}

	result := make([]db.BlogPost, count)
	for i := 0; i < count; i++ {
		result[i] = q.blogPosts[len(q.blogPosts)-1-i]
	}

	return result, nil
}

func (q *LocalQueries) ListBlogPostsByPlatform(_ context.Context, arg db.ListBlogPostsByPlatformParams) ([]db.BlogPost, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Filter by platform, return newest first
	var filtered []db.BlogPost
	for i := len(q.blogPosts) - 1; i >= 0; i-- {
		if q.blogPosts[i].BlogPlatform == arg.BlogPlatform {
			filtered = append(filtered, q.blogPosts[i])
			if len(filtered) >= int(arg.Limit) {
				break
			}
		}
	}

	return filtered, nil
}
