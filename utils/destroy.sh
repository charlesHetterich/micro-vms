#!/usr/bin/env bash

sudo rm -f /tmp/vm1.sock
if [ -n "$1" ]; then
    sudo kill "$1"
    exit 0
fi
sudo pkill firecracker

