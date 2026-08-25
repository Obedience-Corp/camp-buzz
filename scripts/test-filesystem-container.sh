#!/usr/bin/env bash
# Run all real-filesystem tests inside an ephemeral container.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_image="${CAMP_BUZZ_GO_IMAGE:-golang@sha256:298734aec230b5f3e8cee450ce6d7eccc39f1797ba548ee90d57e9803030c6c3}" # 1.25.9-bookworm

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
