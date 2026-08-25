#!/usr/bin/env bash
# Run all real-filesystem tests inside an ephemeral container.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_image="${CAMP_BUZZ_GO_IMAGE:-golang@sha256:47ce5636e9936b2c5cbf708925578ef386b4f8872aec74a67bd13a627d242b19}" # 1.26.2-bookworm

docker run --rm \
  --mount "type=bind,src=${repo_root},dst=/src,readonly" \
  --workdir /work \
  "$go_image" \
  bash -c '
    set -euo pipefail
    cp -a /src/. /work/
    rm -f /work/.git
    go test -tags=integration ./...
  '
