package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	fcSdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/go-openapi/swag"
	"github.com/sirupsen/logrus"
)

const (
	TMP_DIR     = "/tmp/micro-vms"
	BIN         = "/root/micro-vms/bin"
	kernelImage = BIN + "/kernel-unpacked/rootfs/vmlinux"
	rootfsBase  = "/mnt/rootfs-base"
	rootfsImg   = BIN + "/rootfs.ext4"
)

var vmIDs = NewVMIds(TMP_DIR + "/vm_ids.txt")

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fc-manager <launch|snap|launch_snap>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		launch()
	case "launch_snap":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ./fc-manager launch_snap <machine-state-path> <mem-path>")
			os.Exit(1)
		}
		snapPath := os.Args[2]
		memPath := os.Args[3]
		launchSnap(snapPath, memPath)
	case "snap":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./fc-manager connect <vmId>")
			os.Exit(1)
		}
		snap("snapshot.tar", "snapshot.mem", os.Args[2])
	case "connect":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./fc-manager connect <vmId>")
			os.Exit(1)
		}
		connect(os.Args[2])
	case "list":
		records := vmIDs.GetRecords()
		if len(records) == 0 {
			fmt.Println("No VMs running")
			return
		}
		fmt.Println("Running VMs:")
		for _, rec := range records {
			fmt.Printf("ID: %s, PID: %d\n", rec.id, rec.pid)
		}
	case "delete":
		deleteVms(os.Args[2:])
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}

func openTap(id string) error {
	tap := vmIDs.GetTapName(id)
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

func closeTap(id string) error {
	tap := vmIDs.GetTapName(id)
	cmd := exec.Command("ip", "link", "delete", tap)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete tap device %s: %w", tap, err)
	}
	return nil
}

func launchSnap(snapPath string, memPath string) {
	vmId, err := vmIDs.GetAvailableIp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get available IP: %v\n", err)
		return
	}
	ip := vmIDs.GetIp(vmId)
	tap := vmIDs.GetTapName(vmId)
	socketPath := vmIDs.GetSocketPth(vmId)
	mac := vmIDs.GetMacAddress(vmId)
	if err := openTap(vmId); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open tap device: %v\n", err)
		return
	}

	// Create overlay for filesystem
	overlay, err := SetupOverlayWithNetplan(vmId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup overlay: %v\n", err)
		return
	}

	ctx := context.Background()
	_, guestNet, _ := net.ParseCIDR(ip + "/24")

	cfg := fcSdk.Config{
		SocketPath: socketPath,
		Drives: []models.Drive{
			{
				DriveID:      fcSdk.String("rootfs"),
				PathOnHost:   fcSdk.String(rootfsImg),
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
		// Could we set `MachineCfg` here & change amount of mem available?
		NetworkInterfaces: []fcSdk.NetworkInterface{
			{
				StaticConfiguration: &fcSdk.StaticNetworkConfiguration{
					MacAddress:  mac,
					HostDevName: tap,
					IPConfiguration: &fcSdk.IPConfiguration{
						IPAddr:  *guestNet,
						Gateway: net.ParseIP("172.30.0.1"),
						IfName:  "eth0",
					},
				},
			},
		},
	}

	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	cmd := fcSdk.VMCommandBuilder{}.WithSocketPath(socketPath).Build(ctx)
	machine, err := fcSdk.NewMachine(
		ctx, cfg,
		fcSdk.WithLogger(entry),
		fcSdk.WithProcessRunner(cmd),
		fcSdk.WithSnapshot(memPath, snapPath),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = machine.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	pid, _ := machine.PID()
	vmIDs.AddID(vmId, pid)
	fmt.Println("Launched VM", vmId)
}

func deleteVms(vmIds []string) {

	records := vmIDs.GetRecords()
	idSet := make(map[string]struct{}, len(vmIds))
	for _, id := range vmIds {
		idSet[id] = struct{}{}
	}
	// closeTap("2")
	// socketPath := vmIDs.GetSocketPth("2")
	// os.Remove(socketPath)
	var killedIds []string
	for _, record := range records {
		// If vmIds is not empty, only delete those in the list
		if len(vmIds) > 0 {
			if _, ok := idSet[record.id]; !ok {
				continue
			}
		}

		// 1. Kill the VM process
		// if err := syscall.Kill(record.pid, syscall.SIGKILL); err != nil {
		// 	fmt.Fprintf(os.Stderr, "Failed to kill PID %d (VM %s): %v\n", record.pid, record.id, err)
		// 	continue
		// }

		// 2. Remove socket file
		socketPath := vmIDs.GetSocketPth(record.id)
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Failed to remove socket %s: %v\n", socketPath, err)
			// (still count as killed if process is dead)
		}

		// 3. Close tap device
		if err := closeTap(record.id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close tap device for VM %s: %v\n", record.id, err)
			// (still count as killed if process is dead)
		}

		// 4. Unmount overlay and cleanup
		if err := CleanOverlay(record.id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clean overlay for VM %s: %v\n", record.id, err)
			// (still count as killed if process is dead)
		}

		// 4. Mark for removal from table
		killedIds = append(killedIds, record.id)
		fmt.Printf("Deleted VM %s (PID %d)\n", record.id, record.pid)
	}

	if len(killedIds) > 0 {
		if _, err := vmIDs.RmRecords(killedIds); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove records: %v\n", err)
		}
	}
}

func snap(snapPath string, memPath string, vmId string) {
	socketPath := vmIDs.GetSocketPth(vmId)

	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	ctx := context.Background()
	fcClient := fcSdk.NewClient(socketPath, entry, false)

	// Pause the VM
	if _, err := fcClient.PatchVM(ctx, &models.VM{State: swag.String(models.VMStatePaused)}); err != nil {
		panic(fmt.Errorf("pause failed: %w", err))
	}
	fmt.Println("VM is now paused")

	// 5) Take a snapshot (memory + disk)
	snap := &models.SnapshotCreateParams{
		SnapshotPath: swag.String("snapshot.tar"),
		MemFilePath:  swag.String("snapshot.mem"),
	}
	if _, err := fcClient.CreateSnapshot(ctx, snap); err != nil {
		panic(fmt.Errorf("snapshot failed: %w", err))
	}

	// Sleep for 10 seconds
	fmt.Println("Sleeping for 10 seconds...")
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled")
		return
	case <-time.After(10 * time.Second):
	}
	fmt.Println("Done sleeping")

	// 6) Resume the VM
	if _, err := fcClient.PatchVM(ctx, &models.VM{State: swag.String(models.VMStateResumed)}); err != nil {
		panic(fmt.Errorf("resume failed: %w", err))
	}
	fmt.Println("VM resumed")
}

func launch() {
	vmId, err := vmIDs.GetAvailableIp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get available IP: %v\n", err)
		return
	}
	ip := vmIDs.GetIp(vmId)
	tap := vmIDs.GetTapName(vmId)
	socketPath := vmIDs.GetSocketPth(vmId)
	mac := vmIDs.GetMacAddress(vmId)
	if err := openTap(vmId); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open tap device: %v\n", err)
		return
	}

	// Create overlay for filesystem
	// overlay, err := SetupOverlayWithNetplan(vmId)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Failed to setup overlay: %v\n", err)
	// 	return
	// }

	ctx := context.Background()
	_, guestNet, _ := net.ParseCIDR(ip + "/24")
	cfg := fcSdk.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImage,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off overlay_root=ram init=/sbin/overlay-init overlay_id=" + vmId,
		Drives: []models.Drive{
			{
				DriveID:      fcSdk.String("rootfs"),
				PathOnHost:   fcSdk.String(rootfsImg),
				IsRootDevice: fcSdk.Bool(true),
				IsReadOnly:   fcSdk.Bool(false),
			},
			// {
			// 	DriveID:      fcSdk.String("overlayfs"),
			// 	PathOnHost:   fcSdk.String(overlay),
			// 	IsRootDevice: fcSdk.Bool(false),
			// 	IsReadOnly:   fcSdk.Bool(false),
			// },
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  fcSdk.Int64(2),
			MemSizeMib: fcSdk.Int64(256),
			Smt:        fcSdk.Bool(false),
		},
		NetworkInterfaces: []fcSdk.NetworkInterface{
			{
				StaticConfiguration: &fcSdk.StaticNetworkConfiguration{
					MacAddress:  mac,
					HostDevName: tap,
					IPConfiguration: &fcSdk.IPConfiguration{
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

	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	cmd := fcSdk.VMCommandBuilder{}.WithSocketPath(socketPath).Build(ctx)
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
	pid, _ := machine.PID()
	vmIDs.AddID(vmId, pid)
	fmt.Println("Launched VM", vmId)
}

func connect(vmId string) {
	ip := "172.30.0." + vmId
	sshKey := os.Getenv("HOME") + "/.ssh/id_ed25519"
	cmd := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", sshKey,
		"root@"+ip,
	)

	// Attach current terminal to the SSH process for interactivity
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the SSH command
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ssh failed: %v\n", err)
		os.Exit(1)
	}
}

// go build -o fc-manager
// sudo ./fc-manager

// sudo rm /tmp/vm1.sock
// ps aux | grep firecracker
// sudo pkill firecracker

// sudo go run *.go <command>

// mount -t overlay overlay -o lowerdir=/root/micro-vms/bin/rootfs.img,upperdir=/tmp/micro-vms/2/upper,workdir=/tmp/micro-vms/2/work /tmp/micro-vms/2/mount
