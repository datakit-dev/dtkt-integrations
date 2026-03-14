#!/usr/bin/env bash
set -euo pipefail

# Find and stop all integrations in packages/
find . -type f -name "package.dtkt.yaml" | while read -r modfile; do
  dir=$(dirname "$modfile")
  echo "==> Stopping $dir"
  (cd "$dir" && (dtkt i stop || echo "failed to stop $dir"))
done
