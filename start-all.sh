#!/usr/bin/env bash
set -euo pipefail

# Find and start all integrations in packages/
find . -type f -name "package.dtkt.yaml" | while read -r modfile; do
  dir=$(dirname "$modfile")
  echo "==> Starting $dir"
  (cd "$dir" && (dtkt i start -d --local || echo "failed to start $dir"))
done
