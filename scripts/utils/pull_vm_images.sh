#!/bin/sh

BIN=/root/micro-vms/bin
TMP=/tmp/micro-vms
ROOTFS_MOUNT=$TMP/rootfs-base
OVERLAY_INIT=/root/micro-vms/utils/overlay-init

# VM_NUM="$1"
# if [ -z "$VM_NUM" ]; then
#     echo "Usage: $0 <vm_number>"
#     exit 1
# fi

rm -rf $BIN
mkdir -p $BIN
mkdir -p $ROOTFS_MOUNT

skopeo copy docker://ghcr.io/liquidmetal-dev/firecracker-kernel:6.1 oci:$BIN/kernel:latest
skopeo copy docker://ghcr.io/liquidmetal-dev/ubuntu:22.04 oci:$BIN/rootfs:latest

umoci unpack --image $BIN/kernel:latest $BIN/kernel-unpacked
umoci unpack --image $BIN/rootfs:latest $BIN/rootfs-unpacked

# Prepare ext4 rootfs image
fallocate -l 1G $BIN/rootfs.ext4
mkfs.ext4 $BIN/rootfs.ext4
mount $BIN/rootfs.ext4 $ROOTFS_MOUNT
cp -a $BIN/rootfs-unpacked/rootfs/. $ROOTFS_MOUNT/

# Inject SSH key into `authorized_keys`
sudo mkdir -p $ROOTFS_MOUNT/root/.ssh
sudo cp ~/.ssh/id_ed25519.pub $ROOTFS_MOUNT/root/.ssh/authorized_keys
sudo chmod 700 $ROOTFS_MOUNT/root/.ssh
sudo chmod 600 $ROOTFS_MOUNT/root/.ssh/authorized_keys
sudo chown -R root:root $ROOTFS_MOUNT/root/.ssh

# Copy in overlay-init script
sudo cp $OVERLAY_INIT $ROOTFS_MOUNT/sbin/overlay-init
chmod +x $ROOTFS_MOUNT/sbin/overlay-init

cat <<EOF | sudo tee $ROOTFS_MOUNT/etc/netplan/01-netcfg.yaml
network:
    version: 2
    ethernets:
        eth0:
            dhcp4: false
            addresses: [172.30.0.11/24]
            gateway4: 172.30.0.1
            nameservers:
                addresses: [8.8.8.8,1.1.1.1]
EOF
sudo chown root:root $ROOTFS_MOUNT/etc/netplan/01-netcfg.yaml

# mksquashfs $ROOTFS_MOUNT  $BIN/rootfs.ext4 -noappend
sudo umount $ROOTFS_MOUNT
rm -rf $ROOTFS_MOUNT