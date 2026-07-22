ALTER TABLE posts ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;
CREATE INDEX idx_posts_category_id ON posts (category_id);
