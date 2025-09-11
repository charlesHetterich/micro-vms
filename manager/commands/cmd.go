// commands/cmd.go
package commands

import (
	"fmt"
	"io"
	"manager/utils"
	c "manager/utils/constants"
	"os"
	"strings"
	"syscall"
	"time"

	sdkvsock "github.com/firecracker-microvm/firecracker-go-sdk/vsock"
)

func padID(id string) string {
	if len(id) < 3 {
		return fmt.Sprintf("%03s", id)
	}
	return id
}

func tryCloseWrite(conn io.ReadWriteCloser) {
	// Prefer UnixConn.CloseWrite if available.
	type closeWriter interface{ CloseWrite() error }
	if cw, ok := conn.(closeWriter); ok {
		_ = cw.CloseWrite()
		return
	}
	// Fallback: shutdown(SHUT_WR)
	type syscaller interface {
		SyscallConn() (syscall.RawConn, error)
	}
	if sc, ok := conn.(syscaller); ok {
		if rc, err := sc.SyscallConn(); err == nil {
			_ = rc.Control(func(fd uintptr) { _ = syscall.Shutdown(int(fd), syscall.SHUT_WR) })
		}
	}
}

func (a *App) Cmd(id string, argv []string) error {
	id = padID(id)
	if len(argv) == 0 {
		return fmt.Errorf("usage: manager cmd <id> <command ...>")
	}
	meta := utils.VMMetaData(id)
	uds := meta.VsockUDS()
	port := uint32(c.VM_SOCKET_PORT)

	// Be a little more patient than the default 100ms so we don’t race immediately after launch.
	conn, err := sdkvsock.Dial(uds, port,
		sdkvsock.WithRetryTimeout(2*time.Second),
		sdkvsock.WithRetryInterval(50*time.Millisecond),
	)
	if err != nil {
		return fmt.Errorf("dial vsock (uds=%s, port=%d): %w", uds, port, err)
	}
	defer conn.Close()

	// Do we actually have stdin to send? (pipe/file vs TTY)
	fi, _ := os.Stdin.Stat()
	hasIn := fi != nil && (fi.Mode()&os.ModeCharDevice) == 0

	// Send 1-line header + command
	header := "STDIN 0 "
	if hasIn {
		header = "STDIN 1 "
	}
	if _, err := io.WriteString(conn, header+strings.Join(argv, " ")+"\n"); err != nil {
		return err
	}

	// If stdin is piped, stream it then half-close the write side.
	if hasIn {
		_, _ = io.Copy(conn, os.Stdin)
		tryCloseWrite(conn)
	}

	// Read all remote output until the guest closes the connection.
	_, _ = io.Copy(os.Stdout, conn)
	return nil
}
