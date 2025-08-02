BASE_IMG=/root/micro-vms/bin/rootfs.img
BASE_MNT=/mnt/rootfs-base

sudo mkdir -p $BASE_MNT
sudo losetup -fP $BASE_IMG          # attaches to next free loop device (e.g., /dev/loop7)
LOOP_DEV=$(losetup -j $BASE_IMG | cut -d: -f1)
sudo mount -o ro $LOOP_DEV $BASE_MNT


# # EACH TIME U MAKE VM
# VM_ID=42
# VM_DIR=/tmp/vm$VM_ID
# sudo mkdir -p $VM_DIR/{upper,work,merged}
# sudo mount -t overlay overlay \
#   -o lowerdir=$BASE_MNT,upperdir=$VM_DIR/upper,workdir=$VM_DIR/work \
#   $VM_DIR/merged

# # EACH TIME U KILL VM
# sudo umount $VM_DIR/merged

# # Kill the mount stuff
# sudo umount $BASE_MNT
# sudo losetup -d $LOOP_DEV





# dd if=/dev/zero of=overlay.ext4 conv=sparse bs=1M count=1024

# in vm-id dir
# mkfs.ext4 overlay.ext4
