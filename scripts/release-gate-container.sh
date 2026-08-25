#!/usr/bin/env bash
# Run the complete release-readiness gate without writing into the checkout.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
git_dir="$(git -C "$repo_root" rev-parse --absolute-git-dir)"
go_image="${CAMP_BUZZ_GO_IMAGE:-golang@sha256:47ce5636e9936b2c5cbf708925578ef386b4f8872aec74a67bd13a627d242b19}" # 1.26.2-bookworm
gitleaks_image="${CAMP_BUZZ_GITLEAKS_IMAGE:-zricethezav/gitleaks@sha256:cdbb7c955abce02001a9f6c9f602fb195b7fadc1e812065883f695d1eeaba854}" # v8.28.0
goreleaser_image="${CAMP_BUZZ_GORELEASER_IMAGE:-goreleaser/goreleaser@sha256:5be644c8c779677d069b7f50d5e655274c65b5e188c41268abd5b3713c416527}" # v2.15.2
staticcheck_version="${CAMP_BUZZ_STATICCHECK_VERSION:-v0.6.1}"
govulncheck_version="${CAMP_BUZZ_GOVULNCHECK_VERSION:-v1.1.4}"
actionlint_version="${CAMP_BUZZ_ACTIONLINT_VERSION:-v1.7.12}"
before="$(git -C "$repo_root" status --porcelain=v1 --untracked-files=all)"

run_go_gate() {
  docker run --rm \
    --mount "type=bind,src=${repo_root},dst=/src,readonly" \
    --workdir /work \
    -e "STATICCHECK_VERSION=${staticcheck_version}" \
    -e "GOVULNCHECK_VERSION=${govulncheck_version}" \
    -e "ACTIONLINT_VERSION=${actionlint_version}" \
    "$go_image" bash -c '
      set -euo pipefail
      cp -a /src/. /work/
      rm -rf -- /work/.git
      tool_bin="$(go env GOPATH)/bin"
      test -z "$(gofmt -l .)"
      go install "github.com/rhysd/actionlint/cmd/actionlint@${ACTIONLINT_VERSION}"
      "$tool_bin/actionlint" .github/workflows/*.yaml
      go vet -tags=integration ./...
      go test ./...
      go test -race ./...
      go test -tags=integration ./...
      go test -race -tags=integration ./...
      go build -o /tmp/camp-buzz ./cmd/camp-buzz
      go install "honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}"
      "$tool_bin/staticcheck" -tags=integration ./...
      go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
      "$tool_bin/govulncheck" ./...
      go run github.com/goreleaser/goreleaser/v2@v2.15.2 --version >/dev/null
    '
}

run_secret_gates() {
  docker run --rm --entrypoint sh \
    --mount "type=bind,src=${repo_root},dst=/src,readonly" \
    "$gitleaks_image" -c \
    'cp -a /src/. /work && rm -rf -- /work/.git && gitleaks dir --redact /work'
  docker run --rm --entrypoint sh \
    --mount "type=bind,src=${git_dir},dst=/history,readonly" \
    "$gitleaks_image" -c \
    'git clone --quiet /history /repo && gitleaks git --redact /repo'
}

run_release_config_gate() {
  docker run --rm --entrypoint sh \
    --mount "type=bind,src=${repo_root},dst=/src,readonly" \
    "$goreleaser_image" -c \
    'cp -a /src/. /work && rm -rf -- /work/.git && cd /work && goreleaser check'
}

run_go_gate
run_secret_gates
run_release_config_gate

after="$(git -C "$repo_root" status --porcelain=v1 --untracked-files=all)"
if [[ "$after" != "$before" ]]; then
  echo "release gate modified the source checkout" >&2
  diff <(printf '%s\n' "$before") <(printf '%s\n' "$after") || true
  exit 1
fi

echo "RELEASE GATE PASS"
