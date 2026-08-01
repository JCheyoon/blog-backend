package post

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Service holds business rules (slug generation, validation) so handlers
// stay thin and the repository stays a pure data-access layer. This
// separation is also what lets the future MCP server reuse the exact same
// logic instead of duplicating it.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, tag string, category string, publishedOnly bool) ([]Post, error) {
	return s.repo.List(ctx, tag, category, publishedOnly)
}

// ListAll returns posts including drafts. Admin-only: callers must be
// behind the auth middleware.
func (s *Service) ListAll(ctx context.Context, tag string, category string) ([]Post, error) {
	return s.repo.List(ctx, tag, category, false)
}

// GetBySlug is the public read path: drafts are private, so only published
// posts are ever returned.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Post, error) {
	return s.repo.GetBySlug(ctx, slug, true)
}

// GetBySlugAdmin returns any post by slug including drafts. Admin-only:
// callers must be behind the auth middleware.
func (s *Service) GetBySlugAdmin(ctx context.Context, slug string) (*Post, error) {
	return s.repo.GetBySlug(ctx, slug, false)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in CreatePostInput) (*Post, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	if in.Tags == nil {
		in.Tags = []string{}
	}

	p := Post{
		Title:      in.Title,
		Slug:       slugify(in.Title),
		Excerpt:    in.Excerpt,
		Content:    in.Content,
		Tags:       in.Tags,
		CategoryID: in.CategoryID,
		Published:  in.Published,
	}
	return s.repo.Create(ctx, p)
}

func (s *Service) Update(ctx context.Context, id int64, in UpdatePostInput) (*Post, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	p := Post{
		Title:      in.Title,
		Excerpt:    in.Excerpt,
		Content:    in.Content,
		Tags:       in.Tags,
		CategoryID: in.CategoryID,
		Published:  in.Published,
	}
	return s.repo.Update(ctx, id, p)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = slugInvalidChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
