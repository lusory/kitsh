package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/lusory/libkitsune"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"os"
	"strings"
)

// ConsoleCtxKey is a context key for commands invoked in an interactive console.
var ConsoleCtxKey = "console command"

// UnknownFormat is an error about a missing image format.
var UnknownFormat = errors.New("unknown format")

// main is the application entrypoint.
func main() {
	app := &cli.App{
		Name:                 "kitsh",
		Usage:                "A CLI for kitsune's gRPC API",
		EnableBashCompletion: true,
		ExitErrHandler: func(cCtx *cli.Context, err error) {
			if cCtx.Context.Value(ConsoleCtxKey) != nil {
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
				Usage: "should a HTTPS connection be opened instead of HTTP?",
				Value: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "console",
				Aliases: []string{"c", "interactive", "shell"},
				Usage:   "launches an interactive console for issuing commands",
				Action: func(cCtx *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					file, _ := os.Executable()

					for {
						fmt.Print("kitsh> ")
						text, _ := reader.ReadString('\n')
						args := splitBySpace(strings.TrimSuffix(text, "\n"))
						for len(args) > 0 && strings.HasPrefix(args[0], "-") {
							args = args[1:] // remove any global arguments
						}
						if len(args) == 0 {
							continue
						}

						newArgs := append([]string{file, "--target", cCtx.String("target")}, args...)
						if err := cCtx.App.RunContext(context.WithValue(context.Background(), ConsoleCtxKey, args), newArgs); err != nil {
							color.Red("%s", err)
						}
					}
				},
			},
			{
				Name:    "image",
				Aliases: []string{"img", "images", "i"},
				Usage:   "image registry specific actions",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all images",
						Action: func(cCtx *cli.Context) error {
							client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
							if err != nil {
								return err
							}

							images, err := client.ImageRegistry.GetImages(context.Background(), &emptypb.Empty{})
							if err != nil {
								return err
							}

							tbl := table.New("ID", "Format", "Size", "Read-only", "Media type")

							for {
								image, err := images.Recv()
								if err == io.EOF {
									break
								} else if err != nil {
									return err
								}

								tbl.AddRow(image.GetId().GetValue(), image.GetFormat().String(), image.GetSize(), image.GetReadOnly(), image.GetMediaType().String())
							}

							tbl.Print()
							return nil
						},
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
						},
						Action: func(cCtx *cli.Context) error {
							client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
							if err != nil {
								return err
							}

							format, ok := v1.Image_Format_value[strings.ToUpper(cCtx.String("format"))]
							if !ok {
								return UnknownFormat
							}

							oneof, err := client.ImageRegistry.CreateImage(
								context.Background(),
								&v1.CreateImageRequest{
									Format: v1.Image_Format(format),
									Size:   cCtx.Uint64("size"),
								},
							)
							if err != nil {
								return err
							}
							if oneof.GetError() != nil {
								return errors.New(fmt.Sprintf("%s: %s", oneof.GetError().GetType(), oneof.GetError().GetMsg()))
							}

							image := oneof.GetImage()
							tbl := table.New("ID", "Format", "Size", "Read-only", "Media type")
							tbl.AddRow(image.GetId().GetValue(), image.GetFormat().String(), image.GetSize(), image.GetReadOnly(), image.GetMediaType().String())
							tbl.Print()

							return nil
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

// splitBySpace splits the supplied string on spaces, honoring double-quoted substrings.
func splitBySpace(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
}
