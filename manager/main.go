package main

import (
	"fmt"
	"manager/commands"
	"manager/utils"
	"os"
)

const (
	TMP         = "/tmp/micro-vms"
	BIN         = "/root/micro-vms/bin"
	kernelImage = BIN + "/kernel-unpacked/rootfs/vmlinux"
	rootfsBase  = "/mnt/rootfs-base"
	rootfsImg   = BIN + "/rootfs.ext4"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fc-manager <launch|snap|launch_snap>")
		os.Exit(1)
	}

	app := commands.NewApp(utils.NewRecordKeeper(TMP + "/records.json"))

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		app.Launch()
	case "connect":
		fmt.Println("TODO! Implement `connect`")
	case "list":
		app.List(os.Args[2:])
	case "delete":
		app.Delete(os.Args[2:])
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
