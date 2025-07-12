#!/bin/sh
set -euo pipefail

VM_IP="172.30.0.13"
SSH_KEY="$HOME/.ssh/id_ed25519"

# Launch raw vm
VM_ID=$(sh raw_launch.sh)
sh utils/await_connection.sh

scp -o StrictHostKeyChecking=no -i "${SSH_KEY}" \
    ./VM_SCRIPT.sh "root@${VM_IP}:/root/VM_SCRIPT.sh"
ssh -o StrictHostKeyChecking=no -i "${SSH_KEY}" \
    "root@${VM_IP}" 'bash /root/VM_SCRIPT.sh'
