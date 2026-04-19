-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE products (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content    TEXT        NOT NULL,
    metadata   JSONB       NOT NULL DEFAULT '{}'::jsonb,
    embedding  vector(1024) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_embedding_cosine
    ON products
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

ANALYZE products;

-- +goose Down
DROP TABLE IF EXISTS products;
