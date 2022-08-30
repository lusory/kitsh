package main

import (
	"github.com/fatih/color"
	"github.com/lusory/kitsh/handler"
	"github.com/urfave/cli/v2"
	"os"
)

// main is the application entrypoint.
func main() {
	app := &cli.App{
		Name:                 "kitsh",
		Usage:                "A CLI for kitsune's gRPC API",
		EnableBashCompletion: true,
		ExitErrHandler: func(cCtx *cli.Context, err error) {
			if cCtx.Context.Value(handler.ConsoleCtxKey) != nil {
				return // don't exit on interactive console errors
			}

			cli.HandleExitCoder(err)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "target",
				Aliases:  []string{"t", "host"},
				Usage:    "the kitsune target to connect to",
				Required: true,
				EnvVars:  []string{"KITSUNE_TARGET"},
			},
			&cli.BoolFlag{
				Name:  "ssl",
				Usage: "enables SSL (TLS) for gRPC connections",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "no-pretty",
				Usage: "disables pretty-printing of gRPC responses",
				Value: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "console",
				Aliases: []string{"c", "interactive", "shell"},
				Usage:   "launches an interactive console for issuing commands",
				Action:  handler.Console,
			},
			{
				Name:    "image",
				Aliases: []string{"img", "images", "i"},
				Usage:   "image registry specific actions",
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "lists all images",
						Action: handler.ListImages,
					},
					{
						Name:  "create",
						Usage: "creates an image",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "format",
								Aliases:  []string{"f"},
								Usage:    "the image format",
								Required: true,
							},
							&cli.Uint64Flag{
								Name:     "size",
								Aliases:  []string{"s"},
								Usage:    "the image size in bytes, must not be negative",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "data",
								Aliases: []string{"d"},
								Usage:   "the image metadata in JSON (a string-string map)",
								Value:   "{}",
							},
						},
						Action: handler.CreateImage,
					},
					{
						Name:  "delete",
						Usage: "deletes an image",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the image UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.DeleteImage,
					},
					{
						Name:  "metadata",
						Usage: "gets image metadata",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the image UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.GetMetadata,
						Subcommands: []*cli.Command{
							{
								Name:  "set",
								Usage: "sets image metadata",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "data",
										Aliases: []string{"d"},
										Usage:   "the image metadata in JSON (a string-string map)",
									},
								},
								Action: handler.SetMetadata,
							},
							{
								Name:  "clear",
								Usage: "clears image metadata, equivalent to setting '{}' as metadata",
								Flags: []cli.Flag{
									// workaround to not duplicate code from SetMetadata
									&cli.StringFlag{
										Name:   "data",
										Value:  "{}",
										Hidden: true,
									},
								},
								Action: handler.ClearMetadata,
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		color.Red("%s", err)
		os.Exit(1)
	}
}
