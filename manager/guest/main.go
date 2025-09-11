//go:build linux

// guest/main.go
package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	sdkvsock "github.com/firecracker-microvm/firecracker-go-sdk/vsock"
	"github.com/sirupsen/logrus"
)

const port uint32 = 5005

func handle(c io.ReadWriteCloser) {
	defer c.Close()

	br := bufio.NewReader(c)
	line, err := br.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimRight(line, "\r\n")

	// Header: "STDIN 0 <cmd>" or "STDIN 1 <cmd>"
	wantStdin := false
	switch {
	case strings.HasPrefix(line, "STDIN 1 "):
		wantStdin = true
		line = strings.TrimPrefix(line, "STDIN 1 ")
	case strings.HasPrefix(line, "STDIN 0 "):
		line = strings.TrimPrefix(line, "STDIN 0 ")
	default:
		// Back-compat: assume no stdin
	}

	cmd := exec.Command("/bin/sh", "-lc", line)
	cmd.Env = append(os.Environ(),
		"HOME=/root",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	)
	cmd.Dir = "/root" // optional: start in /root
	if wantStdin {
		cmd.Stdin = br
	}
	cmd.Stdout = c
	cmd.Stderr = c

	_ = cmd.Run() // when it returns, defer will Close(), signaling EOF to host
}

func main() {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // quiet; flip to stdout if you want logs
	ctx := context.Background()

	ln, err := sdkvsock.Listener(ctx, logrus.NewEntry(logger), port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}
