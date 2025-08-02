package manager

import (
	"fmt"
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

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		fmt.Println("TODO! Implement `launch`")
	case "connect":
		fmt.Println("TODO! Implement `connect`")
	case "list":
		fmt.Println("TODO! Implement `list`")
	case "delete":
		fmt.Println("TODO! Implement `delete`")
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
