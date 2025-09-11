package scripts

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
)

//go:embed bin/guest-go-init.amd64
var GuestExecd []byte

//go:embed connect.sh
var connectSh []byte

//go:embed pull-root-overlay.sh
var pullRO []byte

// Maps script names -> embedded script blobs
var registry = map[string][]byte{
	"connect": connectSh,
	"pullRO":  pullRO,
}

// Execute script from registry with provided arguments
func Run(name string, args ...string) error {
	data, ok := registry[name]
	if !ok {
		return fmt.Errorf("unknown script %q", name)
	}

	argv := append([]string{"-s"}, args...)
	cmd := exec.Command("sh", argv...)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
