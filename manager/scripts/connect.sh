# Capture IP from args
VM_IP="$1"
if [ -z "$VM_IP" ]; then
    echo "Usage: $0 <IP>"
    exit 1
fi

# Get SSH key
SSH_PUB="$HOME/.ssh/id_ed25519.pub"

# Ping VM until it is ready
START_TIME=$(date +%s)
for i in $(seq 1 20); do
    echo "Attempt $i..."
    ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@${VM_IP}" true && break
    sleep 2
done

# Calculate elapsed time
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
echo "Total time elapsed: ${ELAPSED} seconds"

# actually connect
ssh -o StrictHostKeyChecking=no -i "${SSH_PUB%.*}" "root@${VM_IP}"