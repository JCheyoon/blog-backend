package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("category not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, slug, parent_id, created_at, updated_at
		FROM categories
		ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Category, error) {
	query := `
		SELECT id, name, slug, parent_id, created_at, updated_at
		FROM categories WHERE id = $1`

	var c Category
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &c, nil
}

func (r *Repository) Create(ctx context.Context, c Category) (*Category, error) {
	query := `
		INSERT INTO categories (name, slug, parent_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, c.Name, c.Slug, c.ParentID).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &c, nil
}

func (r *Repository) Update(ctx context.Context, id int64, c Category) (*Category, error) {
	query := `
		UPDATE categories
		SET name = $1, slug = $2, parent_id = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, name, slug, parent_id, created_at, updated_at`

	var updated Category
	err := r.db.QueryRow(ctx, query, c.Name, c.Slug, c.ParentID, id).
		Scan(&updated.ID, &updated.Name, &updated.Slug, &updated.ParentID, &updated.CreatedAt, &updated.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
