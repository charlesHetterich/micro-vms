package main

import (
	"context"
	"fmt"
	"net"
	"os"

	fc "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/go-openapi/swag"
	"github.com/sirupsen/logrus"
)

const (
	BIN         = "/root/micro-vms/bin"
	kernelImage = BIN + "/kernel-unpacked/rootfs/vmlinux"
	rootfsImage = BIN + "/rootfs.img"
	vmID        = "vm1"
	tapDevice   = "tap0"
	guestIP     = "172.30.0.14"
	socketPath  = "/tmp/vm1.sock"
	snapFile    = "/tmp/vm1.snap"
	memFile     = "/tmp/vm1.mem"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fc-demo <launch|snap|launch_snap>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		launch()
	case "snap":
		snap()
	case "launch_snap":
		launchSnap()
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}

func getConfig() fc.Config {
	_, guestNet, _ := net.ParseCIDR(guestIP + "/24")

	return fc.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImage,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off",
		Drives: []models.Drive{
			{
				DriveID:      fc.String("rootfs"),
				PathOnHost:   fc.String(rootfsImage),
				IsRootDevice: fc.Bool(true),
				IsReadOnly:   fc.Bool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  fc.Int64(2),
			MemSizeMib: fc.Int64(256),
			Smt:        fc.Bool(false),
		},
		NetworkInterfaces: []fc.NetworkInterface{
			{
				StaticConfiguration: &fc.StaticNetworkConfiguration{
					MacAddress:  "AA:FC:00:00:00:01",
					HostDevName: tapDevice,
					IPConfiguration: &fc.IPConfiguration{
						IPAddr:  *guestNet,
						Gateway: net.ParseIP("172.30.0.1"),
						IfName:  "eth0",
					},
				},
			},
		},
		LogPath:     "./firecracker.log",
		LogLevel:    "Debug",
		MetricsPath: "./firecracker.metrics",
	}
}

func launch() {
	ctx := context.Background()
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	cfg := getConfig()
	machineOpts := []fc.Opt{
		fc.WithLogger(entry),
	}
	machine, err := fc.NewMachine(ctx, cfg, machineOpts...)
	if err != nil {
		panic(err)
	}
	if err := machine.Start(ctx); err != nil {
		panic(err)
	}
	fmt.Println("VM running with ID:", vmID)
	fmt.Println("Firecracker socket at:", cfg.SocketPath)
	if err := machine.Wait(ctx); err != nil {
		panic(err)
	}
}

func snap() {
	ctx := context.Background()
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	// cfg := getConfig()
	// machineOpts := []fc.Opt{
	// 	fc.WithLogger(entry),
	// }

	// machine, err := fc.NewMachine(ctx, cfg, machineOpts...)
	// if err != nil {
	// 	panic(err)
	// }

	fcClient := fc.NewClient(socketPath, entry, false)
	input := &models.SnapshotCreateParams{
		SnapshotType: "Full",
		MemFilePath:  swag.String(memFile),
		SnapshotPath: swag.String(snapFile),
		// Optional: add 'pause_vm: true' to pause before snapshot
	}

	fmt.Printf("Creating snapshot to %s and %s...\n", snapFile, memFile)
	_, err := fcClient.CreateSnapshot(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("failed to snapshot: %v", err))
	}
	fmt.Println("Snapshot complete.")
}

func launchSnap() {
	ctx := context.Background()
	logger := logrus.New()
	entry := logrus.NewEntry(logger)

	cfg := getConfig()
	// cfg.SnapshotParams = &fc.SnapshotParams{
	// 	SnapshotType: "Full",
	// 	MemFilePath:  memFile,
	// 	SnapshotPath: snapFile,
	// }

	machineOpts := []fc.Opt{
		fc.WithLogger(entry),
	}
	machine, err := fc.NewMachine(ctx, cfg, machineOpts...)
	if err != nil {
		panic(err)
	}

	if err := machine.Start(ctx); err != nil {
		panic(err)
	}
	fmt.Println("Restored VM from snapshot.")
	if err := machine.Wait(ctx); err != nil {
		panic(err)
	}
}
