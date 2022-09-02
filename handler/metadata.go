package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

// MetadatableRegistry is a registry that allows for CRUD operations with metadata.
type MetadatableRegistry interface {
	GetMetadata(ctx context.Context, in *v1.GetMetadataRequest, opts ...grpc.CallOption) (*v1.GetMetadataResponse, error)
	SetMetadata(ctx context.Context, in *v1.SetMetadataRequest, opts ...grpc.CallOption) (*v1.SetMetadataResponse, error)
}

// GetMetadataFunc produces a handler for "metadata" commands.
func GetMetadataFunc(registry MetadatableRegistry) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		id, err := uuid.Parse(cCtx.String("id"))
		if err != nil {
			return err
		}

		meta, err := registry.GetMetadata(
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
			return formatError(meta.GetError())
		}

		data := meta.GetMeta().GetData()
		if data != nil {
			data0, _ := json.Marshal(data)
			fmt.Println(string(data0))
		} else {
			fmt.Println("{}")
		}

		return nil
	}
}

// SetMetadataFunc produces a handler for "metadata set" commands.
func SetMetadataFunc(registry MetadatableRegistry) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		id, err := uuid.Parse(cCtx.String("id"))
		if err != nil {
			return err
		}

		data := make(map[string]string)
		if err := json.Unmarshal([]byte(cCtx.String("data")), &data); err != nil {
			return err
		}

		res, err := registry.SetMetadata(
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
			return formatError(res.GetError())
		}

		return nil
	}
}
