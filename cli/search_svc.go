package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"socrati/cmd/shared"
	"socrati/config"
	"socrati/modules/products"
	"socrati/pkg/embedder"

	urfavecli "github.com/urfave/cli/v2"
)

// RunSearch embeds the query and logs the top-K nearest products.
func RunSearch(c *urfavecli.Context) error {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	q := strings.TrimSpace(c.String("query"))
	if q == "" {
		q = strings.TrimSpace(strings.Join(c.Args().Slice(), " "))
	}
	if q == "" {
		return fmt.Errorf("query is required: pass as positional arg or --query")
	}

	k := c.Int("k")
	if k <= 0 {
		k = 5
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

	start := time.Now()
	vec, err := emb.EmbedAs(ctx, q, embedder.InputTypeQuery)
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}
	embedMs := time.Since(start).Milliseconds()

	repo := products.NewRepository(pool)

	start = time.Now()
	rows, err := repo.SearchTopK(ctx, vec, k)
	if err != nil {
		return err
	}
	searchMs := time.Since(start).Milliseconds()

	log.With(
		"query", q,
		"k", k,
		"embed_ms", embedMs,
		"search_ms", searchMs,
		"hits", len(rows),
	).Info("search complete")

	for i, r := range rows {
		log.With(
			"rank", i+1,
			"distance", r.Distance,
			"id", r.ID,
			"content", r.Content,
		).Info("search hit")
	}
	return nil
}
