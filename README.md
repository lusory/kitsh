# kitsh

A CLI for kitsune's gRPC API, made in Go using libkitsune.

## Installation
You need to have a Go >=1.18 toolchain installed on your system.
```bash
go install github.com/lusory/kitsh@latest
```

## Usage
```
NAME:
   kitsh - A CLI for kitsune's gRPC API

USAGE:
   kitsh [global options] command [command options] [arguments...]

COMMANDS:
   console, c, interactive, shell  launches an interactive console for issuing commands
   image, img, images, i           image registry specific actions
   vm                              virtual machine registry specific actions
   help, h                         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h                              show help (default: false)
   --no-pretty                             disables pretty-printing of gRPC responses (default: false)
   --ssl                                   enables SSL (TLS) for gRPC connections (default: false)
   --target value, -t value, --host value  the kitsune target to connect to [$KITSUNE_TARGET]
```
```
NAME:
   kitsh image - image registry specific actions

USAGE:
   kitsh image command [command options] [arguments...]

COMMANDS:
   list      lists all images
   create    creates an image
   delete    deletes an image
   metadata  gets image metadata
   help, h   Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```
```
NAME:
   kitsh image metadata - gets image metadata

USAGE:
   kitsh image metadata command [command options] [arguments...]

COMMANDS:
   set      sets image metadata
   clear    clears image metadata, equivalent to setting '{}' as metadata
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --id value, -i value  the image UUID (must conform to a v4 UUID)
   --help, -h            show help (default: false)
```
```
NAME:
   kitsh vm - virtual machine registry specific actions

USAGE:
   kitsh vm command [command options] [arguments...]

COMMANDS:
   list      lists all virtual machines
   create    creates a virtual machine
   delete    deletes a virtual machine
   status    queries a virtual machine for status
   images    lists images attached to a virtual machine
   attach    attaches an image to a virtual machine
   detach    detaches an image from a virtual machine
   vnc       launches a HTTP server serving a small VNC viewer
   metadata  gets virtual machine metadata
   help, h   Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```
```
NAME:
   kitsh vm metadata - gets virtual machine metadata

USAGE:
   kitsh vm metadata command [command options] [arguments...]

COMMANDS:
   set      sets virtual machine metadata
   clear    clears virtual machine metadata, equivalent to setting '{}' as metadata
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --id value, -i value  the virtual machine UUID (must conform to a v4 UUID)
   --help, -h            show help (default: false)
```
```
NAME:
   kitsh console - launches an interactive console for issuing commands

USAGE:
   kitsh console [command options] [arguments...]

OPTIONS:
   --help, -h  show help (default: false)
```
