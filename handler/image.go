package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lusory/libkitsune"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"strings"
)

// UnknownFormat is an error about a missing image format.
var UnknownFormat = errors.New("unknown format")

// forEachImages invokes the supplied callback for every image in the supplied stream.
func forEachImages(images v1.ImageRegistryService_GetImagesClient, forEach func(image *v1.Image) error) error {
	for {
		image, err := images.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := forEach(image); err != nil {
			return err
		}
	}

	return nil
}

// ListImages is a handler for the "image list" command.
func ListImages(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	images, err := client.ImageRegistry.GetImages(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	if cCtx.Bool("no-pretty") {
		err = forEachImages(images, func(image *v1.Image) error {
			data, err := json.Marshal(image)
			if err != nil {
				return err
			}

			fmt.Println(string(data))
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		tbl := table.New("ID", "Format", "Size", "Read-only", "Media type")

		err = forEachImages(images, func(image *v1.Image) error {
			tbl.AddRow(image.GetId().GetValue(), image.GetFormat().String(), image.GetSize(), image.GetReadOnly(), image.GetMediaType().String())
			return nil
		})
		if err != nil {
			return err
		}
		tbl.Print()
	}

	return nil
}

// CreateImage is a handler for the "image create" command.
func CreateImage(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	format, ok := v1.Image_Format_value[strings.ToUpper(cCtx.String("format"))]
	if !ok {
		return UnknownFormat
	}

	data := make(map[string]string)
	if err := json.Unmarshal([]byte(cCtx.String("data")), &data); err != nil {
		return err
	}

	oneof, err := client.ImageRegistry.CreateImage(
		context.Background(),
		&v1.CreateImageRequest{
			Format: v1.Image_Format(format),
			Size:   cCtx.Uint64("size"),
			Data: &v1.MetadataMap{
				Data: data,
			},
		},
	)

	if err != nil {
		return err
	}
	if oneof.GetError() != nil {
		return FormatError(oneof.GetError())
	}

	image := oneof.GetImage()

	if cCtx.Bool("no-pretty") {
		data, err := json.Marshal(image)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		return nil
	} else {
		tbl := table.New("ID", "Format", "Size", "Read-only", "Media type", "Metadata")
		tbl.AddRow(
			image.GetId().GetValue(),
			image.GetFormat().String(),
			image.GetSize(),
			image.GetReadOnly(),
			image.GetMediaType().String(),
		)
		tbl.Print()
	}

	return nil
}

// DeleteImage is a handler for the "image delete" command.
func DeleteImage(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	res, err := client.ImageRegistry.DeleteImage(
		context.Background(),
		&v1.DeleteImageRequest{
			Id: &v1.UUID{
				Value: id.String(),
			},
		},
	)

	if err != nil {
		return err
	}
	if res.GetError() != nil {
		return FormatError(res.GetError())
	}

	return nil
}

// GetMetadata is a handler for the "image metadata" command.
func GetMetadata(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	meta, err := client.ImageRegistry.GetMetadata(
		context.Background(),
		&v1.GetMetadataRequest{
			Id: &v1.UUID{
				Value: id.String(),
			},
		},
	)

	if err != nil {
		return err
	}
	if meta.GetError() != nil {
		return FormatError(meta.GetError())
	}

	data, _ := json.Marshal(meta.GetMeta().GetData())
	fmt.Println(string(data))

	return nil
}

// SetMetadata is a handler for the "image metadata set" command.
func SetMetadata(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	data := make(map[string]string)
	if err := json.Unmarshal([]byte(cCtx.String("data")), &data); err != nil {
		return err
	}

	res, err := client.ImageRegistry.SetMetadata(
		context.Background(),
		&v1.SetMetadataRequest{
			Id: &v1.UUID{
				Value: id.String(),
			},
			Meta: &v1.MetadataMap{
				Data: data,
			},
		},
	)

	if err != nil {
		return err
	}
	if res.GetError() != nil {
		return FormatError(res.GetError())
	}

	return nil
}

// ClearMetadata is a handler for the "image metadata clear" command.
func ClearMetadata(cCtx *cli.Context) error {
	if err := cCtx.Set("data", "{}"); err != nil {
		return err
	}
	return SetMetadata(cCtx)
}
