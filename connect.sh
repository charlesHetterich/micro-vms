VM_UID="01JZXN9HB05P96R0A2MP8PY97N"        # ← your new VM’s uid
HOST="0.0.0.0:9090"                        # Flintlockd endpoint
SSH_KEY="$HOME/.ssh/id_ed25519"            # private half of the key in cloud-init

# grab the first ipAddresses[0].address value
IP=$(fl microvm get --host "$HOST" "$VM_UID" -o json \
       | jq -r '.status.networkInterfaces[]?.ipAddresses?[0]?.address' \
       | head -n1)

if [[ -z "$IP" ]]; then
  echo "❌  Flintlock hasn’t recorded an IP for $VM_UID yet."
  echo "    Either DHCP inside the guest failed, or you never assigned a static address."
  exit 1
fi

echo "→ connecting to $IP"


SSH_PUB="$HOME/.ssh/id_ed25519.pub"
ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@172.30.0.10/24" \
    'source ~/.bashrc && bun run index.ts'
