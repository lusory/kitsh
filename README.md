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
   help, h                         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h                              show help (default: false)
   --ssl                                   should a HTTPS connection be opened instead of HTTP? (default: false)
   --target value, -t value, --host value  the kitsune target to connect to [$KITSUNE_TARGET]
```
```
NAME:
   kitsh image - image registry specific actions

USAGE:
   kitsh image command [command options] [arguments...]

COMMANDS:
   list     lists all images
   create   creates an image
   delete   deletes an image
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```
```
NAME:
   kitsh console - launches an interactive console for issuing commands

USAGE:
   kitsh console [command options] [arguments...]

OPTIONS:
   --help, -h  show help (default: false)
```
