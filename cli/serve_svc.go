package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"socrati/cmd/httpapi"
	"socrati/config"
	"socrati/pkg/logger"

	urfavecli "github.com/urfave/cli/v2"
)

// RunServe loads config then starts the root HTTP server (health check + request logging).
func RunServe(c *urfavecli.Context) error {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Logger.Format, cfg.Logger.Level)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := httpapi.NewApp(log)
	return httpapi.Run(ctx, app, addr, log)
}
