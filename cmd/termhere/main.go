package main

import (
	"errors"
	"github.com/guoyk93/termhere"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
)

func envToken() (token string, err error) {
	if token = strings.TrimSpace(os.Getenv("TERMHERE_TOKEN")); token == "" {
		err = errors.New("missing environment variable TERMHERE_TOKEN")
		return
	}
	return
}

func main() {
	app := cli.NewApp()
	app.Usage = "a simple reverse shell tunnel"
	app.Name = "termhere"
	app.Authors = []*cli.Author{
		{
			Name:  "GUO YANKE",
			Email: "hi@guoyk.xyz",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:    "server",
			Usage:   "run as server",
			Aliases: []string{"s"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "listen",
					Usage:   "listen address",
					Aliases: []string{"l"},
					Value:   ":7777",
				},
				&cli.StringFlag{
					Name:  "cert-file",
					Usage: "tls certificate file path",
				},
				&cli.StringFlag{
					Name:  "key-file",
					Usage: "tls certificate key file path",
				},
			},
			Action: func(ctx *cli.Context) (err error) {
				log.Println("running as server")
				var token string
				if token, err = envToken(); err != nil {
					return
				}
				return termhere.RunServer(termhere.ServerOptions{
					Token:    token,
					Listen:   ctx.String("listen"),
					CertFile: ctx.String("cert-file"),
					KeyFile:  ctx.String("key-file"),
				})
			},
		},
		{
			Name:    "client",
			Usage:   "run as client",
			Aliases: []string{"c"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "server",
					Usage:    "server address",
					Aliases:  []string{"s"},
					Value:    "",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "ca-file",
					Usage: "TLS ca file for server",
				},
				&cli.BoolFlag{
					Name:    "insecure",
					Usage:   "skip tls verification",
					Aliases: []string{"k"},
				},
			},
			Action: func(ctx *cli.Context) (err error) {
				log.Println("running as client")
				var token string
				if token, err = envToken(); err != nil {
					return
				}
				command := ctx.Args().Slice()
				if len(command) == 0 {
					if shell := strings.TrimSpace(os.Getenv("SHELL")); shell != "" {
						command = []string{shell}
					} else {
						err = errors.New("missing command")
						return
					}
				}
				return termhere.RunClient(termhere.ClientOptions{
					Token:    token,
					Server:   ctx.String("server"),
					Command:  command,
					CAFile:   ctx.String("ca-file"),
					Insecure: ctx.Bool("insecure"),
				})
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}
}
