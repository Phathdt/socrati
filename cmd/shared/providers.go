package shared

import (
	"context"
	"fmt"
	"time"

	"socrati/config"
	"socrati/pkg/embedder"
	"socrati/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitLogger creates and initializes the application logger.
func InitLogger(cfg *config.Config) logger.Logger {
	return logger.New(cfg.Logger.Format, cfg.Logger.Level)
}

// InitDatabase opens a pgxpool to Postgres. Caller owns Close().
func InitDatabase(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// InitEmbedder builds the configured embedder provider. Today only "voyage" is
// supported; add more branches here when new providers land.
func InitEmbedder(cfg *config.Config, log logger.Logger) (*embedder.VoyageEmbedder, error) {
	switch cfg.Embedder.Provider {
	case "voyage":
		return embedder.NewVoyage(embedder.VoyageConfig{
			APIKey:     cfg.Embedder.APIKey,
			Model:      cfg.Embedder.Model,
			BaseURL:    cfg.Embedder.BaseURL,
			Timeout:    time.Duration(cfg.Embedder.TimeoutMS) * time.Millisecond,
			MaxRetries: cfg.Embedder.MaxRetries,
			MaxChars:   cfg.Embedder.MaxChars,
		}, log)
	default:
		return nil, fmt.Errorf("unsupported embedder provider: %q", cfg.Embedder.Provider)
	}
}
