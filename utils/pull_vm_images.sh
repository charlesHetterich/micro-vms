skopeo copy docker://ghcr.io/liquidmetal-dev/firecracker-kernel:6.1 oci:kernel:latest
skopeo copy docker://ghcr.io/liquidmetal-dev/ubuntu:22.04 oci:rootfs:latest

umoci unpack --image kernel:latest kernel-unpacked
umoci unpack --image rootfs:latest rootfs-unpacked
