#!/bin/bash

# Configure Network
echo -e "nameserver 8.8.8.8\nnameserver 1.1.1.1" > /etc/resolv.conf

apt-get update
apt-get install -y unzip

# Install Bun
curl -fsSL https://bun.sh/install | bash
source ~/.bashrc

bun init --yes
bun add chalk
bun index.ts