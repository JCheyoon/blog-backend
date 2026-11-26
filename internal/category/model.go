package category

import "time"

type Category struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	ParentID  *int64     `json:"parentId,omitempty"`
	Children  []Category `json:"children,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
