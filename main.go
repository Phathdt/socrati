// @title Go Template API
// @version 1.0
// @description E-commerce API with products, orders, payments, promotions, shipping, and auth
// @host localhost:4000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" (without quotes)
package main

import (
	"log"
	"os"

	mycli "socrati/cli"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "socrati",
		Usage: "Go template application",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Start the root HTTP server (health check + request logging)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
				},
				Action: mycli.RunServe,
			},
			{
				Name:      "embed",
				Usage:     "Embed a piece of text via the configured embedder",
				ArgsUsage: "[text...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
					&cli.StringFlag{
						Name:  "text",
						Usage: "Text to embed (overrides positional args)",
					},
				},
				Action: mycli.RunEmbed,
			},
			{
				Name:  "migrate",
				Usage: "Run database migrations (goose)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "config.yml"},
				},
				Subcommands: []*cli.Command{
					{Name: "up", Usage: "Apply pending migrations", Action: mycli.RunMigrateUp},
					{Name: "down", Usage: "Rollback last migration", Action: mycli.RunMigrateDown},
					{Name: "status", Usage: "Show migration status", Action: mycli.RunMigrateStatus},
				},
			},
			{
				Name:  "seed",
				Usage: "Embed sample products and insert into Postgres",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "config.yml"},
					&cli.StringFlag{Name: "file", Usage: "Path to JSON seed file (optional)"},
				},
				Action: mycli.RunSeed,
			},
			{
				Name:  "bulk-seed",
				Usage: "Insert N synthetic rows (random unit vectors) for perf testing",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "config.yml"},
					&cli.IntFlag{Name: "count", Aliases: []string{"n"}, Value: 1000, Usage: "Rows to insert"},
				},
				Action: mycli.RunBulkSeed,
			},
			{
				Name:      "search",
				Usage:     "Embed a query and return top-K similar products",
				ArgsUsage: "[query...]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "config.yml"},
					&cli.StringFlag{Name: "query", Aliases: []string{"q"}, Usage: "Search query"},
					&cli.IntFlag{Name: "k", Value: 5, Usage: "Top K results"},
				},
				Action: mycli.RunSearch,
			},
		},
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
