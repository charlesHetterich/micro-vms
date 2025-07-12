#!/usr/bin/env bash
set -euo pipefail

NAME="bun-demo-$(date +%s)"
NS="default"
ADDR="0.0.0.0:9090"
CLOUDINIT="/root/dothome/scratch/cloudinit.yaml"

SSH_PUB="$HOME/.ssh/id_ed25519.pub"
GUEST_IP="172.30.0.13/24"



# Build the YAML in the CC variable
CC=$(cat <<'EOF'
#cloud-config
runcmd:
  - curl -fsSL https://bun.sh/install | bash
  - ln -s /root/.bun/bin/bun /usr/local/bin/bun
  - bun add chalk
  - echo 'import chalk from "chalk"; console.log(chalk.red("Hello from cloud-init!"))' > /root/index.ts
EOF
)

# Base64-encode in one line, strip newlines (-w0; if your base64
# lacks -w use `tr -d '\n'` afterwards)
UD=$(printf '%s' "$CC" | base64 -w0)

VM_ID=$(fl microvm create \
  --host "$ADDR" \
  --name "$NAME" \
  --namespace "$NS" \
  --vcpu 2 \
  --memory 1024 \
  --kernel-image ghcr.io/liquidmetal-dev/firecracker-kernel:6.1 \
  --root-image   ghcr.io/liquidmetal-dev/ubuntu:22.04 \
  --network-interface net0:tap::${GUEST_IP} \
  --metadata-from-file user-data="${CLOUDINIT}" \
  2>&1 | grep -oE '[0-9A-Z]{26}' | head -n1)
#   --metadata-hostname "$NAME" \
#   --metadata-ssh-key-file "$SSH_PUB" \

echo "Waiting for VM $VM_ID to boot"
sleep 2
sudo journalctl -u flintlockd | grep $VM_ID -n -B2 -A4 



# ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@172.30.0.13" 'source ~/.bashrc && bun run index.ts'
cat /var/lib/flintlock/vm/default/$NAME/$VM_ID/metadata.json
