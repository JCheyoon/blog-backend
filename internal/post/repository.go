package post

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("post not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, tag string, category string, publishedOnly bool) ([]Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.excerpt, p.content, p.tags, p.category_id, p.published, p.created_at, p.updated_at
		FROM posts p
		WHERE ($1 = '' OR $1 = ANY(p.tags))
		  AND ($2 = '' OR p.category_id = ANY(
		        ARRAY(SELECT id FROM categories WHERE slug = $2 OR parent_id = (SELECT id FROM categories WHERE slug = $2))
		      ))
		  AND (NOT $3 OR p.published = true)
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(ctx, query, tag, category, publishedOnly)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.Tags, &p.CategoryID, &p.Published, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// GetBySlug fetches a post by slug. When publishedOnly is true, draft
// (unpublished) posts are excluded so private posts never leak through
// public routes.
func (r *Repository) GetBySlug(ctx context.Context, slug string, publishedOnly bool) (*Post, error) {
	query := `
		SELECT id, title, slug, excerpt, content, tags, category_id, published, created_at, updated_at
		FROM posts
		WHERE slug = $1
		  AND (NOT $2 OR published = true)`

	var p Post
	err := r.db.QueryRow(ctx, query, slug, publishedOnly).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.Tags, &p.CategoryID, &p.Published, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get post by slug: %w", err)
	}
	return &p, nil
}
