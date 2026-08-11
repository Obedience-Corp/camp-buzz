#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
completions_dir="${repo_root}/completions"
tmp_binary="${completions_dir}/.camp-buzz-tmp"

cleanup() {
  rm -f "${tmp_binary}"
}

mkdir -p "${completions_dir}"
trap cleanup EXIT

echo "Building temporary camp-buzz binary for completion generation..."
(
  cd "${repo_root}"
  go build -o "${tmp_binary}" ./cmd/camp-buzz
)

echo "Generating completions..."
(
  cd "${repo_root}"
  "${tmp_binary}" completion bash > "${completions_dir}/camp-buzz.bash"
  "${tmp_binary}" completion zsh > "${completions_dir}/_camp-buzz"
  "${tmp_binary}" completion fish > "${completions_dir}/camp-buzz.fish"
)

echo "Completions generated in ${completions_dir}/"
