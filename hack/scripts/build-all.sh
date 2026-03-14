#!/usr/bin/env bash
set -euo pipefail

# Find and build all integrations in packages/
find . -type f -name "package.dtkt.yaml" | while read -r modfile; do
  dir=$(dirname "$modfile")
  echo "==> Building $dir"
  (cd "$dir" && (dtkt i build || echo "failed to build $dir"))
done
