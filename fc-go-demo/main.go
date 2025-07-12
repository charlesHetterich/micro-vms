package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	fc "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	logger := logrus.New()
	entry := logrus.NewEntry(logger)

	kernelImage := "/path/to/kernel-image"
	rootfsImage := "/path/to/rootfs-image"

	// vmName := "bun-demo"
	vmID := "vm1"
	tapDevice := "tap0"
	guestIP := "172.30.0.14"
	_, guestNet, _ := net.ParseCIDR(guestIP + "/24")

	// 1. Ensure tap interface exists:
	// exec.Command("ip", "tuntap", "add", "dev", tapDevice, "mode", "tap").Run()
	// exec.Command("ip", "addr", "add", "172.30.0.1/24", "dev", tapDevice).Run()
	// exec.Command("ip", "link", "set", tapDevice, "up").Run()

	// 2. Firecracker Config
	cfg := fc.Config{
		SocketPath:      filepath.Join(os.TempDir(), vmID+".sock"),
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
