package category

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Tree returns categories as a nested tree (top-level + children).
func (s *Service) Tree(ctx context.Context) ([]Category, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	childMap := make(map[int64][]Category)
	var roots []Category

	for _, c := range all {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			childMap[*c.ParentID] = append(childMap[*c.ParentID], c)
		}
	}

	for i := range roots {
		roots[i].Children = childMap[roots[i].ID]
	}

	return roots, nil
}

// FlatList returns all categories as a flat list (no nesting).
func (s *Service) FlatList(ctx context.Context) ([]Category, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, in CreateCategoryInput) (*Category, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	c := Category{
		Name:     name,
		Slug:     slugify(name),
		ParentID: in.ParentID,
	}
	return s.repo.Create(ctx, c)
}

func (s *Service) Update(ctx context.Context, id int64, in UpdateCategoryInput) (*Category, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
	}

	c := Category{
		Name:     name,
		Slug:     slugify(name),
		ParentID: existing.ParentID,
	}
	if in.ParentID != nil {
		c.ParentID = in.ParentID
	}

	return s.repo.Update(ctx, id, c)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugInvalidChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
