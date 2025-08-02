#!/bin/sh

VM_IP="$1"
if [ -z "$VM_IP" ]; then
    echo "Usage: $0 <VM_IP>"
    exit 1
fi

SSH_PUB="$HOME/.ssh/id_ed25519.pub"
START_TIME=$(date +%s)

# Wait for the VM to be ready
for i in $(seq 1 20); do
    echo "Attempt $i..."
    ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@${VM_IP}" true && break
    sleep 2
done

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
echo "Total time elapsed: ${ELAPSED} seconds"