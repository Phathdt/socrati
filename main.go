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
		},
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
