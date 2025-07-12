#!/usr/bin/env bash
# Run with: sudo bash setup.sh
set -euo pipefail

### ── tunables ────────────────────────────────────────────────────────────
FL_VERSION="main"            # git ref for flintlock; change to a tag if you want
FL_CLI_VERSION="0.3.0"       # version of the fl CLI to install
BRIDGE="lmbr0"               # bridge created for the micro-VMs
GRPC_ADDR="0.0.0.0:9090"     # Flintlockd gRPC endpoint (insecure)
TAP_NAME=tap0                # name of the tap device created by Flintlockd
###########################################################################

msg(){ printf '\n\033[1;32m==> %s\033[0m\n' "$*"; }

[[ $EUID -eq 0 ]] || { echo "Run as root (sudo bash setup.sh)"; exit 1; }

msg "Installing containerd + helpers"
apt-get update
apt-get install -y containerd jq iproute2 iptables ca-certificates curl git

# Install nerdctl, used to login to ghcr.io & fetch images
curl -L \
  https://github.com/containerd/nerdctl/releases/download/v2.1.3/nerdctl-2.1.3-linux-amd64.tar.gz \
  | sudo tar -xz -C /usr/local/bin nerdctl
# INSTEAD OF nerdctl:
sudo apt-get install -y skopeo umoci



systemctl enable --now containerd

# Setup NETWORKING
msg "Creating bridge $BRIDGE"
ip link add "$BRIDGE" type bridge 2>/dev/null || true
ip addr add 172.30.0.1/24 dev "$BRIDGE" 2>/dev/null || true
ip link set "$BRIDGE" up
sysctl -w net.ipv4.ip_forward=1
# tap setup START (scracth)
sudo ip tuntap add $TAP_NAME mode tap
sudo ip link set $TAP_NAME master $BRIDGE up
modprobe tun
echo tun | tee /etc/modules-load.d/tun.conf >/dev/null
# make bridge public
sudo iptables -I FORWARD 1 -i lmbr0 -o eth0 -j ACCEPT
sudo iptables -I FORWARD 1 -i eth0  -o lmbr0 -m state --state RELATED,ESTABLISHED -j ACCEPT
# something about making the connection perminent
sudo apt-get install -y iptables-persistent
sudo sh -c 'iptables-save > /etc/iptables/rules.v4'
echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/99-flintlock.conf
sudo sysctl --system
# END



msg "Fetching flintlock provision helper"
mkdir -p /opt/flintlock && cd /opt/flintlock
curl -fsSL https://raw.githubusercontent.com/liquidmetal-dev/flintlock/"$FL_VERSION"/hack/scripts/provision.sh -o provision.sh
chmod +x provision.sh

msg "Installing Firecracker"
./provision.sh firecracker       # latest upstream binary

msg "Installing flintlockd + fl CLI"
./provision.sh flintlock \
    --dev --insecure \
    --bridge "$BRIDGE" \
    --grpc-address "$GRPC_ADDR"
curl -L "https://github.com/liquidmetal-dev/fl/releases/download/v${FL_CLI_VERSION}/fl_${FL_CLI_VERSION}_linux_amd64.tar.gz" \
  | tar -xz -C /usr/local/bin fl

systemctl enable --now flintlockd       # started with the args above

msg "Setup complete — flintlockd is listening on $GRPC_ADDR"

# INSIDE OF VPM, CONFIGURE NETWORK
# ip route add default via 172.30.0.1 dev eth1
# resolvectl dns eth1 8.8.8.8 1.1.1.1
