package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"socrati/cmd/shared"
	"socrati/config"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	urfavecli "github.com/urfave/cli/v2"
)

const (
	bulkVectorDim = 1024
	bulkBatchSize = 500
)

// RunBulkSeed inserts N synthetic products with random unit vectors.
// Use to benchmark pgvector index performance without hitting the embedder
// provider's rate limit. Runs ANALYZE at the end so the planner picks up
// fresh row counts.
func RunBulkSeed(c *urfavecli.Context) error {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	count := c.Int("count")
	if count <= 0 {
		count = 1000
	}

	log := shared.InitLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := shared.InitDatabase(ctx, cfg.Database.URI)
	if err != nil {
		return err
	}
	defer pool.Close()

	metadata := json.RawMessage(`{"synthetic":true}`)
	start := time.Now()

	inserted := 0
	for offset := 0; offset < count; offset += bulkBatchSize {
		end := offset + bulkBatchSize
		if end > count {
			end = count
		}

		batch := &pgx.Batch{}
		for i := offset; i < end; i++ {
			content := fmt.Sprintf("synthetic product #%06d", i)
			vec := randomUnitVector(bulkVectorDim)
			batch.Queue(
				`INSERT INTO products (content, metadata, embedding) VALUES ($1, $2, $3)`,
				content, metadata, pgvector.NewVector(vec),
			)
		}

		br := pool.SendBatch(ctx, batch)
		for i := offset; i < end; i++ {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("batch exec at row %d: %w", i, err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close batch: %w", err)
		}

		inserted = end
		log.With("inserted", inserted, "total", count).Info("bulk seed progress")
	}

	// Retrain ivfflat centroids on the full corpus — otherwise the index
	// built on the initial rows will route queries to the wrong list and
	// recall craters (approximate search miss).
	if _, err := pool.Exec(ctx, "REINDEX INDEX idx_products_embedding_cosine"); err != nil {
		return fmt.Errorf("reindex: %w", err)
	}
	if _, err := pool.Exec(ctx, "ANALYZE products"); err != nil {
		return fmt.Errorf("analyze: %w", err)
	}

	log.With(
		"count", inserted,
		"elapsed_ms", time.Since(start).Milliseconds(),
	).Info("bulk seed done")
	return nil
}

// randomUnitVector returns a uniformly random unit vector. Cosine distance is
// meaningful only on normalized vectors, so we normalize here.
func randomUnitVector(dim int) []float32 {
	v := make([]float32, dim)
	var norm float64
	for i := range v {
		f := rand.Float64()*2 - 1 // [-1, 1)
		v[i] = float32(f)
		norm += f * f
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		v[0] = 1
		return v
	}
	inv := float32(1.0 / norm)
	for i := range v {
		v[i] *= inv
	}
	return v
}
