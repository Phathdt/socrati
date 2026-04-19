package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"socrati/cmd/shared"
	"socrati/config"

	urfavecli "github.com/urfave/cli/v2"
)

// RunEmbed embeds the text passed as a positional arg (or via --text) and
// prints a short summary. Primary use: smoke test for the Embedder abstraction.
func RunEmbed(c *urfavecli.Context) error {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	text := strings.TrimSpace(c.String("text"))
	if text == "" {
		text = strings.TrimSpace(strings.Join(c.Args().Slice(), " "))
	}
	if text == "" {
		return errors.New("text is required: pass as positional arg or --text")
	}

	log := shared.InitLogger(cfg)

	emb, err := shared.InitEmbedder(cfg, log)
	if err != nil {
		return fmt.Errorf("build embedder: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	vec, err := emb.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}

	log.With(
		"model", cfg.Embedder.Model,
		"dim", len(vec),
		"preview", formatPreview(vec, 5),
	).Info("embed complete")
	return nil
}

func formatPreview(vec []float32, n int) string {
	if n > len(vec) {
		n = len(vec)
	}
	parts := make([]string, 0, n)
	for _, v := range vec[:n] {
		parts = append(parts, fmt.Sprintf("%.5f", v))
	}
	return strings.Join(parts, ", ")
}
