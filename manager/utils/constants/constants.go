package utils

const (
	// TODO! make sure these are [1] portable and [2] coordinate correctly w/ scripts
	BIN          = "/root/.micro-vm/bin"
	TMP          = "/tmp/micro-vms"
	KERNEL_IMAGE = BIN + "/kernel-unpacked/rootfs/vmlinux"
	ROOTFS_BASE  = "/mnt/rootfs-base"
	ROOTFS_IMG   = BIN + "/rootfs.ext4"
)
