#!/usr/bin/env bash
set -euo pipefail

NAME="bun-demo-$(date +%s)"
NS="default"
ADDR="0.0.0.0:9090"
CLOUDINIT="/root/dothome/scratch/cloudinit.yaml"

SSH_PUB="$HOME/.ssh/id_ed25519.pub"
GUEST_IP="172.30.0.13/24"

VM_ID=$(fl microvm create \
  --host "$ADDR" \
  --name "$NAME" \
  --namespace "$NS" \
  --vcpu 2 \
  --memory 1024 \
  --kernel-image ghcr.io/liquidmetal-dev/firecracker-kernel:6.1 \
  --root-image   ghcr.io/liquidmetal-dev/ubuntu:22.04 \
  --network-interface net0:tap::${GUEST_IP} \
  --metadata-hostname "$NAME" \
  --metadata-ssh-key-file "$SSH_PUB" \
  2>&1 | grep -oE '[0-9A-Z]{26}' | head -n1)
echo $VM_ID