ID=$1
OVERLAY_FN=/tmp/micro-vms/$ID/overlay.ext4
dd if=/dev/zero of=$OVERLAY_FN conv=sparse bs=1M count=1024
mkfs.ext4 $OVERLAY_FN