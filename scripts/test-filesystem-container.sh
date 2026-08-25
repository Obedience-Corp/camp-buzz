#!/usr/bin/env bash
# Run all real-filesystem tests inside an ephemeral container.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_image="${CAMP_BUZZ_GO_IMAGE:-golang@sha256:4557cf171e3cdf5053a298d5171b1a5f5734d920260c25f22c79e94760eb2084}" # 1.25.8-bookworm

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
