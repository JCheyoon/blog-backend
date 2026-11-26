package post

import "context"

// Service decides the read policy (published posts only) so handlers stay
// thin and the repository stays a pure data-access layer.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, tag string, category string, publishedOnly bool) ([]Post, error) {
	return s.repo.List(ctx, tag, category, publishedOnly)
}

// GetBySlug is the public read path: drafts are private, so only published
// posts are ever returned.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Post, error) {
	return s.repo.GetBySlug(ctx, slug, true)
}
