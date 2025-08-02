#!/bin/sh

VM_NUM="$1"
if [ -z "$VM_NUM" ]; then
    echo "Usage: $0 <vm_number>"
    exit 1
fi


SSH_PUB="$HOME/.ssh/id_ed25519.pub"
VM_IP="172.30.0.${VM_NUM}"

sh ./utils/await_connection.sh "$VM_IP"

# actually connect
ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@${VM_IP}"