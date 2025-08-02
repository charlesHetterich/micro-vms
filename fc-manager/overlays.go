package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Unmount overlay and cleanup (call on VM delete)
func CleanOverlay(vmId string) error {
	overlayDir := filepath.Join(TMP_DIR, vmId)
	if err := os.RemoveAll(overlayDir); err != nil {
		return fmt.Errorf("failed to remove overlay directory: %w", err)
	}
	return nil
}

func SetupOverlayWithNetplan(vmId string) (string, error) {
	overlayDir := filepath.Join(TMP_DIR, vmId)
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

//
//
//
//
//
//
//

// Creates overlay, mounts it, writes netplan config for the VM's IP
func _SetupOverlayWithNetplan(vmId string) (string, error) {
	vmIP := vmIDs.GetIp(vmId)
	overlayDir := filepath.Join(TMP_DIR, vmId)
	upperDir := filepath.Join(overlayDir, "upper")
	workDir := filepath.Join(overlayDir, "work")
	mountPoint := filepath.Join(overlayDir, "mount")
	for _, d := range []string{upperDir, workDir, mountPoint} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", err
		}
	}

	// 2. Mount overlay (lowerdir is your base rootfs, mounted at /mnt/rootfs-base)
	mountCmd := exec.Command(
		"mount", "-t", "overlay", "overlay",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", rootfsBase, upperDir, workDir),
		mountPoint,
	)
	if err := mountCmd.Run(); err != nil {
		return "", fmt.Errorf("overlay mount failed: %w", err)
	}

	// 3. Write netplan config in the overlay with proper IP
	netplanPath := filepath.Join(mountPoint, "etc/netplan/01-netcfg.yaml")
	netplan := fmt.Sprintf(`
network:
    version: 2
    ethernets:
        eth0:
            dhcp4: false
            addresses: [%s/24]
            gateway4: 172.30.0.1
            nameservers:
                addresses: [8.8.8.8,1.1.1.1]
`, vmIP)

	if err := os.WriteFile(netplanPath, []byte(netplan), 0644); err != nil {
		return "", fmt.Errorf("write netplan: %w", err)
	}
	if err := os.Chown(netplanPath, 0, 0); err != nil { // root:root
		return "", err
	}
	exec.Command("chown", "root:root", netplanPath).Run()
	return mountPoint, nil // Use as rootfs for the VM
}
