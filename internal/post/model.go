package post

import "time"

type Post struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Excerpt    string    `json:"excerpt"`
	Content    string    `json:"content"`
	Tags       []string  `json:"tags"`
	CategoryID *int64    `json:"categoryId,omitempty"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CreatePostInput struct {
	Title      string   `json:"title"`
	Excerpt    string   `json:"excerpt"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	CategoryID *int64   `json:"categoryId,omitempty"`
	Published  bool     `json:"published"`
}

type UpdatePostInput struct {
	Title      string   `json:"title"`
	Excerpt    string   `json:"excerpt"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	CategoryID *int64   `json:"categoryId,omitempty"`
	Published  bool     `json:"published"`
}
