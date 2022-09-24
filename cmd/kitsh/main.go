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
								Usage:    "the image size in bytes, must be positive",
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
						Action: handler.GetImageMetadata,
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
								Action: handler.SetImageMetadata,
							},
							{
								Name:  "clear",
								Usage: "clears image metadata, equivalent to setting '{}' as metadata",
								Flags: []cli.Flag{
									// workaround to not duplicate code from SetImageMetadata
									&cli.StringFlag{
										Name:   "data",
										Value:  "{}",
										Hidden: true,
									},
								},
								Action: handler.ClearImageMetadata,
							},
						},
					},
				},
			},
			{
				Name:  "vm",
				Usage: "virtual machine registry specific actions",
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "lists all virtual machines",
						Action: handler.ListVirtualMachines,
					},
					{
						Name:  "create",
						Usage: "creates a virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "arch",
								Aliases:  []string{"a"},
								Usage:    "the virtual machine architecture",
								Required: true,
							},
							&cli.Uint64Flag{
								Name:     "memory",
								Aliases:  []string{"m"},
								Usage:    "the virtual machine RAM size in megabytes, must be positive",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "data",
								Aliases: []string{"d"},
								Usage:   "the virtual machine metadata in JSON (a string-string map)",
								Value:   "{}",
							},
						},
						Action: handler.CreateVirtualMachine,
					},
					{
						Name:  "delete",
						Usage: "deletes a virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.DeleteVirtualMachine,
					},
					{
						Name:  "status",
						Usage: "queries a virtual machine for status",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.GetStatus,
					},
					{
						Name:  "images",
						Usage: "lists images attached to a virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.Images,
					},
					{
						Name:  "attach",
						Usage: "attaches an image to a virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "image",
								Usage:    "the image UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.AttachImage,
					},
					{
						Name:  "detach",
						Usage: "detaches an image from a virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "image",
								Usage:    "the image UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.DetachImage,
					},
					{
						Name:  "vnc",
						Usage: "launches a HTTP server serving a small VNC viewer",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "http-host",
								Usage: "the host that the VNC viewer should be served on",
								Value: ":8000",
							},
						},
						Action: handler.VNC,
					},
					{
						Name:  "power",
						Usage: "sends a power command to the virtual machine",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "action",
								Aliases:  []string{"a"},
								Usage:    "the power action (poweron, poweroff, reset)",
								Required: true,
							},
						},
					},
					{
						Name:  "metadata",
						Usage: "gets virtual machine metadata",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Usage:    "the virtual machine UUID (must conform to a v4 UUID)",
								Required: true,
							},
						},
						Action: handler.GetVmMetadata,
						Subcommands: []*cli.Command{
							{
								Name:  "set",
								Usage: "sets virtual machine metadata",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "data",
										Aliases: []string{"d"},
										Usage:   "the virtual machine metadata in JSON (a string-string map)",
									},
								},
								Action: handler.SetVmMetadata,
							},
							{
								Name:  "clear",
								Usage: "clears virtual machine metadata, equivalent to setting '{}' as metadata",
								Flags: []cli.Flag{
									// workaround to not duplicate code from SetVmMetadata
									&cli.StringFlag{
										Name:   "data",
										Value:  "{}",
										Hidden: true,
									},
								},
								Action: handler.ClearVmMetadata,
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
