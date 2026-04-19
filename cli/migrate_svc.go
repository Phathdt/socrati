package cli

import (
	"database/sql"
	"fmt"

	"socrati/cmd/shared"
	"socrati/config"
	"socrati/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	urfavecli "github.com/urfave/cli/v2"
)

const migrationsDir = "migrations"

func openMigrateDB(c *urfavecli.Context) (*sql.DB, logger.Logger, error) {
	cfg, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	db, err := sql.Open("pgx", cfg.Database.URI)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	return db, shared.InitLogger(cfg), nil
}

// RunMigrateUp applies all pending migrations.
func RunMigrateUp(c *urfavecli.Context) error {
	db, log, err := openMigrateDB(c)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	log.Info("migrations applied")
	return nil
}

// RunMigrateDown rolls back the most recent migration.
func RunMigrateDown(c *urfavecli.Context) error {
	db, log, err := openMigrateDB(c)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	log.Info("rollback complete")
	return nil
}

// RunMigrateStatus prints migration status.
func RunMigrateStatus(c *urfavecli.Context) error {
	db, _, err := openMigrateDB(c)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Status(db, migrationsDir); err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}
	return nil
}
