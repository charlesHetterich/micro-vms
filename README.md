# micro-vms
The primary project is the cli tool in `manager`. Run a command with `go run . <cmd>`.

## When Closing
- Delete Record from `VMIds`


## Core Components

### Bridge

### Overlay

### SSH entry

### Custom Mac & IP (via base image)

### VM record

## Compiling

Build guest binary with
```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ./scripts/bin/guest-go-init.amd64 ./guest
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ./scripts/bin/guest-go-init.arm64 ./guest
```
