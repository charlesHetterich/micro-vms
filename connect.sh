#!/bin/sh

SSH_PUB="$HOME/.ssh/id_ed25519.pub"
VM_IP="172.30.0.13"

sh ./utils/await_connection.sh

# actually connect
ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@${VM_IP}"