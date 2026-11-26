package category

import "context"

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
