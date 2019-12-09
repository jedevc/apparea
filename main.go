package main

import (
	"log"
	"os"

	"github.com/jedevc/AppArea/tunnel"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "apparea",
		Usage: "reverse proxying server over ssh!",
		Commands: []*cli.Command{
			{
				Name:  "setup",
				Usage: "automatically initialize config files",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "force overwrite of existing config",
					},
				},
				Action: func(c *cli.Context) error {
					force := c.Bool("force")
					err := tunnel.InitializeConfigs(force)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "run",
				Usage: "run the server",
				Action: func(c *cli.Context) error {
					config, err := tunnel.LoadConfig()
					if err != nil {
						return err
					}

					server := &tunnel.Server{
						Config: &config,
					}
					sessions := server.Run("0.0.0.0:2200")
					for range sessions {
					}

					select {}
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
