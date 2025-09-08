package commands

import (
	"context"
	"fmt"
	"manager/utils"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	c "manager/utils/constants"

	fcSdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
)

func (a *App) Launch() error {
	// 1] Add record (no PID)
	id, err := a.Records.Add(-1)
	if err != nil {
		return fmt.Errorf("`launch` command failed: %v", err)
	}
	meta := utils.VMMetaData(id)
	if err := openTap(meta); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open tap device: %v\n", err)
		return err
	}
	fmt.Printf("VM IP: %s\n", meta.IP())
	ip := net.ParseIP(meta.IP()) // e.g. "172.30.0.7"
	mask := net.CIDRMask(24, 32)

	// _, guestNet, _ := net.ParseCIDR(meta.IP() + "/24")
	fmt.Printf("VM IP: %s\n", net.IPNet{IP: ip, Mask: mask})

	overlay, err := SetupOverlayWithNetplan(id)
	if err != nil {
		return err
	}

	cfg := fcSdk.Config{
		SocketPath:      meta.SocketPth(),
		KernelImagePath: c.KERNEL_IMAGE,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda ro overlay_root=vdb init=/sbin/overlay-init overlay_id=" + id,
		// KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off overlay_root=ram init=/sbin/init overlay_id=" + id,
		// KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off init=/sbin/init",
		// KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init",
		Drives: []models.Drive{
			{
				DriveID:      fcSdk.String("rootfs"),
				PathOnHost:   fcSdk.String(c.ROOTFS_IMG),
				IsRootDevice: fcSdk.Bool(true),
				IsReadOnly:   fcSdk.Bool(true),
			},
			{
				DriveID:      fcSdk.String("overlayfs"),
				PathOnHost:   fcSdk.String(overlay),
				IsRootDevice: fcSdk.Bool(false),
				IsReadOnly:   fcSdk.Bool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  fcSdk.Int64(2),
			MemSizeMib: fcSdk.Int64(256),
			Smt:        fcSdk.Bool(false),
		},
		NetworkInterfaces: []fcSdk.NetworkInterface{
			{
				StaticConfiguration: &fcSdk.StaticNetworkConfiguration{
					MacAddress:  meta.MacAddress(),
					HostDevName: meta.TapName(),
					IPConfiguration: &fcSdk.IPConfiguration{
						IPAddr:  net.IPNet{IP: ip, Mask: mask},
						Gateway: net.ParseIP("172.30.0.1"),
						IfName:  "eth0",
					},
				},
			},
		},
		LogLevel:    "Debug",
		LogPath:     c.BIN + "/firecracker.log",
		MetricsPath: c.BIN + "/firecracker.metrics",
	}

	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	ctx := context.Background()
	cmd := fcSdk.VMCommandBuilder{}.
		WithSocketPath(meta.SocketPth()).
		Build(ctx)
	// cmd := fcSdk.VMCommandBuilder{}.WithSocketPath(meta.SocketPth()).Build(ctx)

	machine, err := fcSdk.NewMachine(
		ctx, cfg,
		fcSdk.WithLogger(entry),
		fcSdk.WithProcessRunner(cmd),
	)
	if err != nil {
		panic(err)
	}
	if err := machine.Start(ctx); err != nil {
		panic(err)
	}
	// pid, _ := machine.PID()
	pid, err := machine.PID()
	if err != nil {
		return fmt.Errorf("get PID: %w", err)
	}

	if err := a.Records.Update(id, pid); err != nil {
		return fmt.Errorf("record PID: %w", err)
	}

	return nil
}

func openTap(meta utils.VMMetaData) error {
	// tap := vmIDs.GetTapName(id)
	tap := meta.TapName()
	cmd := exec.Command("ip", "tuntap", "add", tap, "mode", "tap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tap device %s: %w", tap, err)
	}
	linkCmd := exec.Command("ip", "link", "set", tap, "master", "lmbr0", "up")
	if err := linkCmd.Run(); err != nil {
		return fmt.Errorf("failed to set tap device %s up and attach to bridge: %w", tap, err)
	}
	return nil
}

func SetupOverlayWithNetplan(vmId string) (string, error) {
	overlayDir := filepath.Join(c.TMP, vmId)
	overlayPath := filepath.Join(overlayDir, "overlay.ext4")
	if err := os.MkdirAll(overlayDir, 0755); err != nil {
		return "", err
	}

	// dd if=/dev/zero of=$OVERLAY_FN conv=sparse bs=1M count=1024
	if err := exec.Command("dd", "if=/dev/zero", "of="+overlayPath, "conv=sparse", "bs=1M", "count=1024").Run(); err != nil {
		return "", fmt.Errorf("failed to create overlay file: %w", err)
	}
	if err := exec.Command("mkfs.ext4", overlayPath).Run(); err != nil {
		return "", fmt.Errorf("failed to create ext4 filesystem: %w", err)
	}

	return overlayPath, nil
}
