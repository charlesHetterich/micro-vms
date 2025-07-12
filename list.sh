#!/bin/sh
set -euo pipefail

HOST="0.0.0.0:9090"

if [ "${1:-}" = "--full" ] || [ "${1:-}" = "-f" ]; then
    fl microvm get --host "$HOST"
else
    IDS=$(fl microvm get --host "$HOST" 2>&1 | awk '/^[0-9A-Z]{26}/ {print $1}')
    echo "$IDS"
fi
