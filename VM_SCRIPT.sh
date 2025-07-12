#!/bin/bash
set -euo pipefail

# Configure Network
ip route add default via 172.30.0.1 dev eth1
resolvectl dns eth1 8.8.8.8 1.1.1.1

apt-get update
apt-get install -y unzip

# Install Bun
curl -fsSL https://bun.sh/install | bash
source ~/.bashrc

bun init --yes
bun add chalk
bun index.ts