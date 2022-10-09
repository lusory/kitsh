package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lusory/kitsh"
	"github.com/lusory/libkitsune"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"
)

// UnknownArchitecture is an error about an unknown architecture.
var UnknownArchitecture = errors.New("unknown architecture")

// InvalidRAMSize is an error about an invalid RAM memory size (size must be positive).
var InvalidRAMSize = errors.New("invalid ram size")

// UnknownPowerAction is an error about an unknown power action.
var UnknownPowerAction = errors.New("unknown power action")

// NoOpenWebSocket is an error about an opened WebSocket not being found in a virtual machine's VNC servers.
var NoOpenWebSocket = errors.New("no open websocket found")

// forEachVms invokes the supplied callback for every virtual machine in the supplied stream.
func forEachVms(vms v1.VirtualMachineRegistryService_GetVirtualMachinesClient, forEach func(image *v1.VirtualMachine) error) error {
	for {
		vm, err := vms.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := forEach(vm); err != nil {
			return err
		}
	}

	return nil
}

// ListVirtualMachines is a handler for the "vm list" command.
func ListVirtualMachines(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	vms, err := client.VmRegistry.GetVirtualMachines(cCtx.Context, &emptypb.Empty{})
	if err != nil {
		return err
	}

	if cCtx.Bool("no-pretty") {
		err = forEachVms(vms, func(vm *v1.VirtualMachine) error {
			data, err := json.Marshal(vm)
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
		tbl := table.New("ID", "Architecture", "Memory size")

		err = forEachVms(vms, func(vm *v1.VirtualMachine) error {
			tbl.AddRow(vm.GetId().GetValue(), vm.GetArch(), vm.GetMemorySize())
			return nil
		})
		if err != nil {
			return err
		}
		tbl.Print()
	}

	return nil
}

// CreateVirtualMachine is a handler for the "vm create" command.
func CreateVirtualMachine(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	mem := cCtx.Uint64("memory")
	if mem <= 0 {
		return InvalidRAMSize
	}

	arch, ok := v1.Architecture_value[strings.ToUpper(cCtx.String("arch"))]
	if !ok {
		return UnknownArchitecture
	}

	data := make(map[string]string)
	if err := json.Unmarshal([]byte(cCtx.String("data")), &data); err != nil {
		return err
	}

	oneof, err := client.VmRegistry.CreateVirtualMachine(
		cCtx.Context,
		&v1.CreateVirtualMachineRequest{
			Arch:       v1.Architecture(arch),
			MemorySize: mem,
			Data: &v1.MetadataMap{
				Data: data,
			},
		},
	)

	if err != nil {
		return err
	}
	if oneof.GetError() != nil {
		return formatError(oneof.GetError())
	}

	vm := oneof.GetMachine()

	if cCtx.Bool("no-pretty") {
		data, err := json.Marshal(vm)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		return nil
	} else {
		tbl := table.New("ID", "Architecture", "Memory size")
		tbl.AddRow(
			vm.GetId().GetValue(),
			vm.GetArch(),
			vm.GetMemorySize(),
		)
		tbl.Print()
	}

	return nil
}

// DeleteVirtualMachine is a handler for the "vm delete" command.
func DeleteVirtualMachine(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.DeleteVirtualMachine(
		cCtx.Context,
		&v1.DeleteVirtualMachineRequest{
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

	return nil
}

// GetStatus is a handler for the "vm status" command.
func GetStatus(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.IsAlive(
		cCtx.Context,
		&v1.IsAliveRequest{
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

	if cCtx.Bool("no-pretty") {
		fmt.Println(res.GetAlive())
	} else {
		fmt.Print("Status: ")
		if res.GetAlive() {
			color.Green("Running")
		} else {
			color.Red("Stopped")
		}
	}
	return nil
}

// Images is a handler for the "vm images" command.
func Images(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	images, err := client.VmRegistry.GetAttachedImages(
		cCtx.Context,
		&v1.GetAttachedImagesRequest{
			Id: &v1.UUID{
				Value: id.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	if images.GetError() != nil {
		return formatError(images.GetError())
	}

	if cCtx.Bool("no-pretty") {
		for _, image := range images.GetImages() {
			fmt.Println(image.GetValue())
		}
	} else {
		tbl := table.New("Image ID")

		for _, image := range images.GetImages() {
			tbl.AddRow(image.GetValue())
		}

		tbl.Print()
	}

	return nil
}

// AttachImage is a handler for the "vm attach" command.
func AttachImage(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	image, err := uuid.Parse(cCtx.String("image"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.AttachImage(
		cCtx.Context,
		&v1.AttachImageRequest{
			Machine: &v1.UUID{
				Value: id.String(),
			},
			Image: &v1.UUID{
				Value: image.String(),
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

// DetachImage is a handler for the "vm detach" command.
func DetachImage(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	image, err := uuid.Parse(cCtx.String("image"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.DetachImage(
		cCtx.Context,
		&v1.DetachImageRequest{
			Machine: &v1.UUID{
				Value: id.String(),
			},
			Image: &v1.UUID{
				Value: image.String(),
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

	res, err := client.VmRegistry.GetVNCServers(
		cCtx.Context,
		&v1.GetVNCServersRequest{
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

	sock := findWebSocket(res.GetServers())
	if sock == nil {
		return NoOpenWebSocket
	}

	httpHost := cCtx.String("http-host")
	if strings.HasPrefix(httpHost, ":") { // add localhost prefix if only port is defined
		httpHost = "localhost" + httpHost
	}

	vncHost := cCtx.String("target")
	if strings.Contains(vncHost, ":") { // remove port
		vncHost = strings.SplitN(vncHost, ":", 2)[0]
	}

	//goland:noinspection HttpUrlsUsage - no TLS certificate support
	url := fmt.Sprintf("http://%s/?host=%s&port=%d&path=", httpHost, vncHost, sock.GetPort())
	if cCtx.Bool("no-pretty") {
		fmt.Println(url)
	} else {
		color.Green("A VNC viewer is running: %s", url)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)

	srv, err := serveVNC(httpHost, wg)
	if err != nil {
		return err
	}

	if !cCtx.Bool("no-pretty") {
		color.Yellow("Press 'Enter' to stop the HTTP server.")
	}
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	defer wg.Wait()
	if err := srv.Shutdown(cCtx.Context); err != nil {
		return err
	}

	return nil
}

// Power is a handler for the "vm power" command.
func Power(cCtx *cli.Context) error {
	client, err := libkitsune.NewOrCachedKitsuneClient(cCtx.String("target"), cCtx.Bool("ssl"))
	if err != nil {
		return err
	}

	action, ok := v1.PowerAction_value[strings.ToUpper(cCtx.String("action"))]
	if !ok {
		return UnknownPowerAction
	}

	id, err := uuid.Parse(cCtx.String("id"))
	if err != nil {
		return err
	}

	res, err := client.VmRegistry.SendPowerAction(
		cCtx.Context,
		&v1.SendPowerActionRequest{
			Machine: &v1.UUID{
				Value: id.String(),
			},
			Action: v1.PowerAction(action),
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
func serveVNC(target string, wg *sync.WaitGroup) (*http.Server, error) {
	novncFs, err := fs.Sub(kitsh.NoVNCEmbed, "noVNC")
	if err != nil {
		return &http.Server{}, err
	}

	indexPage, err := kitsh.NoVNCEmbed.ReadFile("noVNC/vnc_lite.html")
	if err != nil {
		return &http.Server{}, err
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

	srv := &http.Server{Addr: target, Handler: r}

	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			color.Red("http server errored (%s)", err.Error())
		}
	}()

	return srv, nil
}

// findWebSocket tries to find the first open WebSocket, returns nil if no WebSocket is open.
func findWebSocket(servers []*v1.VNCServer) *v1.VNCServerSocket {
	for _, vncServer := range servers {
		for _, sock := range vncServer.GetSockets() {
			if sock.GetIsWebSocket() {
				return sock
			}
		}
	}
	return nil
}
