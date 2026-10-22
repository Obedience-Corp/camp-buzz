#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
common="$(git -C "$root" rev-parse --path-format=absolute --git-common-dir)"
gitdir="$(git -C "$root" rev-parse --absolute-git-dir)"
relative="${gitdir#"$common"}"
docker run --rm -v "$root:/src:ro" -v "$common:/history:ro" \
  -v camp-buzz-candidate-gomod:/go/pkg/mod -v camp-buzz-candidate-build:/root/.cache/go-build \
  -e "CANDIDATE_GIT_DIR=/history$relative" -w /work golang:1.26-bookworm bash -euc '
    cp -a /src/. /work/
    rm -f /work/.git
    git config --global --add safe.directory /work
    # Git context is scoped to release commands, never filesystem fixture tests.
    export GIT_DIR="$CANDIDATE_GIT_DIR" GIT_WORK_TREE=/work
    CANDIDATE_VERSION=0.1.3 go run github.com/goreleaser/goreleaser/v2@v2.15.2 release --snapshot --clean
    go run ./internal/releaseutil artifacts --dir dist --tag v0.1.3
    WINDOWS_CANDIDATE=true CANDIDATE_VERSION=0.1.3 go run github.com/goreleaser/goreleaser/v2@v2.15.2 release --snapshot --clean
    go run ./internal/releaseutil artifacts --dir dist --tag v0.1.3 --windows-candidate
    arch="$(go env GOARCH)"
    if [ "$arch" = amd64 ]; then arch=x86_64; fi
    mkdir /native-proof
    tar -xzf "dist/camp-buzz-0.1.3-linux-$arch.tar.gz" -C /native-proof
    /native-proof/camp-buzz version
    /native-proof/camp-buzz version | grep "camp-buzz 0.1.3"
  '
