package handler

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lusory/kitsh"
	"github.com/lusory/libkitsune"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
	"github.com/urfave/cli/v2"
	"io/fs"
	"net/http"
	"strings"
)

// VNC is a handler for the "vm vnc" command.
func VNC(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.GetVNCServer(
		context.Background(),
		&v1.GetVNCServerRequest{
			Id: &v1.UUID{
				Value: id.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	if res.GetError() != nil {
		return formatError(res.GetError())
	}

	vncServer := res.GetServer()
	httpHost := cCtx.String("http-host")
	if strings.HasPrefix(httpHost, ":") { // add localhost prefix if only port is defined
		httpHost = "localhost" + httpHost
	}

	vncHost := cCtx.String("target")
	if strings.Contains(vncHost, ":") { // remove port
		vncHost = strings.SplitN(vncHost, ":", 2)[0]
	}

	//goland:noinspection HttpUrlsUsage - no TLS certificate support
	url := fmt.Sprintf("http://%s/?host=%s&port=%d&path=", httpHost, vncHost, vncServer.GetWebSocketPort())
	if cCtx.Bool("no-pretty") {
		fmt.Println(url)
	} else {
		color.Green("A VNC viewer is running: %s", url)
	}
	return serveVNC(httpHost)
}

// GetVmMetadata is a handler for the "vm metadata" command.
func GetVmMetadata(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	return GetMetadataFunc(client.VmRegistry)(cCtx)
}

// SetVmMetadata is a handler for the "vm metadata set" command.
func SetVmMetadata(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	return SetMetadataFunc(client.VmRegistry)(cCtx)
}

// ClearVmMetadata is a handler for the "vm metadata clear" command.
func ClearVmMetadata(cCtx *cli.Context) error {
	if err := cCtx.Set("data", "{}"); err != nil {
		return err
	}
	return SetImageMetadata(cCtx)
}

// serveVNC starts an HTTP server serving a small noVNC application.
func serveVNC(target string) error {
	novncFs, err := fs.Sub(kitsh.NoVNCEmbed, "noVNC")
	if err != nil {
		return err
	}

	indexPage, err := kitsh.NoVNCEmbed.ReadFile("noVNC/vnc_lite.html")
	if err != nil {
		return err
	}

	novncServ := http.FileServer(http.FS(novncFs))

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexPage)
	})
	r.Get("/core/*", novncServ.ServeHTTP)
	r.Get("/vendor/*", novncServ.ServeHTTP)

	return http.ListenAndServe(target, r)
}
