package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"socrati/cmd/shared"
	"socrati/config"
	"socrati/modules/products"
	"socrati/pkg/embedder"

	urfavecli "github.com/urfave/cli/v2"
)

// defaultSeedFile is the bundled product catalog.
const defaultSeedFile = "seed/products.json"

type seedItem struct {
	Content  string          `json:"content"`
	Metadata json.RawMessage `json:"metadata"`
}

// RunSeed embeds each product in the seed file and inserts rows into Postgres.
func RunSeed(c *urfavecli.Context) error {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	path := c.String("file")
	if path == "" {
		path = defaultSeedFile
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read seed file %q: %w", path, err)
	}
	var items []seedItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("parse seed file: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("seed file has no items")
	}

	log := shared.InitLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := shared.InitDatabase(ctx, cfg.Database.URI)
	if err != nil {
		return err
	}
	defer pool.Close()

	emb, err := shared.InitEmbedder(cfg, log)
	if err != nil {
		return fmt.Errorf("build embedder: %w", err)
	}

	repo := products.NewRepository(pool)

	if err := repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("truncate products: %w", err)
	}

	texts := make([]string, len(items))
	for i, it := range items {
		texts[i] = it.Content
	}

	vecs, err := emb.EmbedBatch(ctx, texts, embedder.InputTypeDocument)
	if err != nil {
		return fmt.Errorf("batch embed: %w", err)
	}

	for i, it := range items {
		id, err := repo.Insert(ctx, it.Content, it.Metadata, vecs[i])
		if err != nil {
			return fmt.Errorf("insert item %d: %w", i, err)
		}
		log.With("id", id, "content", truncate(it.Content, 60)).Info("seed inserted")
	}

	log.With("count", len(items)).Info("seed complete")
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
