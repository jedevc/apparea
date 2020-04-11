package main

import (
	"log"
	"os"

	"github.com/jedevc/AppArea/config"
	"github.com/jedevc/AppArea/tunnel"
	"github.com/urfave/cli/v2"
)

const defaultHostname = "apparea.dev"
const defaultBindAddress = ":2222"

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
					err := config.InitializeConfigs(force)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "serve",
				Usage: "run the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "bind",
						Usage:       "address to bind to",
						DefaultText: defaultBindAddress,
					},
					&cli.StringFlag{
						Name:        "hostname",
						Usage:       "hostname of the server",
						DefaultText: defaultHostname,
					},
				},
				Action: func(c *cli.Context) error {
					if len(c.String("bind")) == 0 {
						err := c.Set("bind", defaultBindAddress)
						if err != nil {
							panic(err)
						}
					}
					if len(c.String("hostname")) == 0 {
						err := c.Set("hostname", defaultHostname)
						if err != nil {
							panic(err)
						}
					}

					config, err := config.LoadConfig()
					if err != nil {
						return err
					}

					server := &tunnel.Server{
						Config:   &config,
						Hostname: c.String("hostname"),
					}
					sessions := server.Run(c.String("bind"))
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
