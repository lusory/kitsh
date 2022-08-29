package main

import (
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

var UnknownFormat = errors.New("unknown format")

// main is the application entrypoint.
func main() {
	app := &cli.App{
		Name:                 "kitsh",
		Usage:                "A CLI for kitsune's gRPC API",
		EnableBashCompletion: true,
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
				Name:    "image",
				Aliases: []string{"img", "images", "i"},
				Usage:   "image registry specific actions",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all images",
						Action: func(cCtx *cli.Context) error {
							client, err := libkitsune.NewKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
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
							client, err := libkitsune.NewKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
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
