// Package products provides persistence for embedded product rows backed by
// Postgres + pgvector. Repository wraps sqlc-generated queries for simple
// writes and a raw pgx query for cosine top-K search.
package products

import (
	"context"
	"encoding/json"
	"fmt"

	"socrati/modules/products/gen"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// Chunk is the domain-level result returned by SearchTopK.
type Chunk struct {
	ID       string          `json:"id"`
	Content  string          `json:"content"`
	Metadata json.RawMessage `json:"metadata"`
	Distance float32         `json:"distance"` // cosine distance (0 = identical)
}

// Repository persists and searches products.
type Repository struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewRepository wires a Repository around an existing pgxpool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: gen.New(pool)}
}

// Insert stores a product + its embedding and returns the generated id.
func (r *Repository) Insert(
	ctx context.Context, content string, metadata json.RawMessage, embedding []float32,
) (string, error) {
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	row, err := r.q.InsertProduct(ctx, gen.InsertProductParams{
		Content:   content,
		Metadata:  metadata,
		Embedding: pgvector.NewVector(embedding),
	})
	if err != nil {
		return "", fmt.Errorf("insert product: %w", err)
	}
	return row.ID, nil
}

// Count returns total products in the table.
func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.q.CountProducts(ctx)
}

// DeleteAll wipes the products table. Test/seed helper.
func (r *Repository) DeleteAll(ctx context.Context) error {
	return r.q.DeleteAllProducts(ctx)
}

// SearchTopK returns the k nearest neighbours to v by cosine distance.
// Written with raw pgx because sqlc cannot parse the `<=>` operator.
// ivfflat is approximate: higher probes = better recall, slower. 10 is a
// pragmatic default (lists=100 → ~10% scan).
func (r *Repository) SearchTopK(ctx context.Context, v []float32, k int) ([]Chunk, error) {
	if k <= 0 {
		k = 5
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // read-only, rollback is always fine

	if _, err := tx.Exec(ctx, "SET LOCAL ivfflat.probes = 10"); err != nil {
		return nil, fmt.Errorf("set probes: %w", err)
	}

	const sql = `
		SELECT id::text, content, metadata, embedding <=> $1 AS distance
		FROM products
		ORDER BY embedding <=> $1
		LIMIT $2
	`
	rows, err := tx.Query(ctx, sql, pgvector.NewVector(v), k)
	if err != nil {
		return nil, fmt.Errorf("search top-k: %w", err)
	}
	defer rows.Close()

	out := make([]Chunk, 0, k)
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.Content, &c.Metadata, &c.Distance); err != nil {
			return nil, fmt.Errorf("scan top-k row: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top-k rows: %w", err)
	}
	return out, nil
}
