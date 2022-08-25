# kitsh

A CLI for kitsune's gRPC API, made in Go using libkitsune.

# Installation
You need to have a Go >=1.18 toolchain installed on your system.
```bash
go install github.com/lusory/kitsh@latest
```

# Usage
```
-command string
    the command to be interpreted (optional, used for one-liners without the console)
-interactive
    should an interactive console be opened? (default false)
-ssl
    should an HTTPS connection be established? (default false)
-target string
    the kitsune gRPC target
```

The interactive console works in a `<registry>.<method> [data]` format. **Example:**
* Currently available registries: `img` (ImageRegistryService) and `vm` (VirtualMachineRegistryService)
* Methods are from the specified registry struct, see [libkitsune](https://github.com/lusory/libkitsune/blob/master/proto/kitsune/proto/v1/image_grpc.pb.go#L25) for those.
* `data` is the JSON version of the protobuf request, defaults to an empty struct
```
kitsh> vm.GetVirtualMachines

kitsh> vm.FindVirtualMachine {"id":{"value":"add21b54-fc45-404d-b11d-222763e40172"}}
```