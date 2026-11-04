#!/usr/bin/env bash
# Build the static agentmem binary and install it to ~/.local/bin (or $1).
set -euo pipefail
cd "$(dirname "$0")/.."
dest="${1:-$HOME/.local/bin}"
mkdir -p "$dest"
CGO_ENABLED=0 GOFLAGS=-mod=vendor go build -o "$dest/agentmem" ./cmd/agentmem
echo "installed $dest/agentmem"
case ":$PATH:" in *":$dest:"*) ;; *) echo "note: $dest is not on PATH" ;; esac
