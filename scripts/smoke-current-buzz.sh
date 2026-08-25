#!/usr/bin/env bash
# Exercise camp-buzz against real Buzz binaries and a disposable local relay.
set -euo pipefail

: "${BUZZ_BIN_DIR:?set BUZZ_BIN_DIR to the directory containing buzz and camp-buzz}"
: "${BUZZ_RELAY_URL:?set BUZZ_RELAY_URL to a disposable local relay}"
: "${CAMP_BUZZ_CONTAINER_SMOKE:?run this filesystem-mutating smoke only in a disposable container}"
if [[ "$CAMP_BUZZ_CONTAINER_SMOKE" != "1" ]]; then
  echo "CAMP_BUZZ_CONTAINER_SMOKE must equal 1" >&2
  exit 1
fi

readonly local_relay_url="http://localhost:3030"
readonly bind_relay_url="ws://localhost:3030"
if [[ "$BUZZ_RELAY_URL" != "$local_relay_url" ]]; then
  echo "BUZZ_RELAY_URL must be $local_relay_url for this local-only smoke" >&2
  exit 1
fi

smoke_root="$(mktemp -d)"
cleanup() {
  unset BUZZ_PRIVATE_KEY
  rm -rf -- "$smoke_root"
}
trap cleanup EXIT

export PATH="${BUZZ_BIN_DIR}:${PATH}"
export HOME="$smoke_root/home"
export CAMP_ROOT="$smoke_root/campaign"
readonly output_dir="$smoke_root/output"
mkdir -p "$HOME" "$CAMP_ROOT/.campaign" "$output_dir"

generate_private_key() {
  local candidate
  while true; do
    candidate="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"
    if python3 - "$candidate" <<'PY'
import sys

key = int(sys.argv[1], 16)
order = int("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
raise SystemExit(0 if 0 < key < order else 1)
PY
    then
      printf '%s' "$candidate"
      return
    fi
  done
}

BUZZ_PRIVATE_KEY="$(generate_private_key)"
export BUZZ_PRIVATE_KEY

channel_json="$(buzz channels create \
  --name camp-buzz-CB0001-disposable \
  --type stream \
  --visibility open \
  --ttl 300)"
channel="$(printf '%s' "$channel_json" | python3 -c \
  'import json,sys; print(json.load(sys.stdin)["channel_id"])')"
test -n "$channel"

camp-buzz bind \
  --channel "$channel" \
  --relay "$bind_relay_url" \
  --festival CB0001 \
  --festival-path festivals/active/camp-buzz-first-public-release-CB0001 \
  >"$output_dir/bind.out"
grep -q "wrote" "$output_dir/bind.out"
grep -q "relay_url: http://localhost:3030" \
  "$CAMP_ROOT/.campaign/integrations/buzz.yaml"

doctor="$(camp-buzz doctor)"
printf '%s\n' "$doctor" | grep -q "status: ready"
printf '%s\n' "$doctor" | grep -q "BUZZ_PRIVATE_KEY: set"

camp-buzz post \
  -m "CB0001 flag path" \
  --task CB0001:P001.S03.T02 \
  --gate pass \
  >"$output_dir/flag.out"
printf "CB0001 stdin path\n" | camp-buzz post \
  --task CB0001:P001.S03.T02-stdin \
  --gate pending \
  >"$output_dir/stdin.out"

hook_yaml="$(camp-buzz hook-install)"
printf '%s\n' "$hook_yaml" | grep -q "command: camp buzz post --from-hook"
printf '%s\n' "$hook_yaml" | grep -q "fail: open"
camp-buzz post \
  --from-hook \
  --task CB0001:P001.S03.T02-hook \
  --gate n/a \
  >"$output_dir/hook.out"

buzz messages get --channel "$channel" --limit 10 >"$output_dir/messages.json"
python3 - "$output_dir/messages.json" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as messages_file:
    contents = {event["content"] for event in json.load(messages_file)}

path = "festivals/active/camp-buzz-first-public-release-CB0001"
expected = {
    "CB0001 flag path\n\n---\nfestival: CB0001\ntask: CB0001:P001.S03.T02\n"
    f"path: {path}\ngate: pass\n---\n",
    "CB0001 stdin path\n\n---\nfestival: CB0001\ntask: CB0001:P001.S03.T02-stdin\n"
    f"path: {path}\ngate: pending\n---\n",
    "Festival status update\n\n---\nfestival: CB0001\ntask: CB0001:P001.S03.T02-hook\n"
    f"path: {path}\ngate: n/a\n---\n",
}
missing = expected - contents
if missing:
    raise SystemExit(f"missing expected messages: {len(missing)}")
print("readback_messages=3")
print("exact_footers=PASS")
PY

expect_failure() {
  local output="$1"
  local expected="$2"
  shift 2
  if "$@" >"$output" 2>&1; then
    echo "expected command to fail: $*" >&2
    return 1
  fi
  grep -q "$expected" "$output"
}

expect_failure "$output_dir/fail-key.out" "BUZZ_PRIVATE_KEY is not set" \
  env -u BUZZ_PRIVATE_KEY camp-buzz post -m rejected
expect_failure "$output_dir/fail-channel.out" "channel id must be a UUID" \
  camp-buzz post -m rejected --channel not-a-uuid
expect_failure "$output_dir/fail-empty.out" "message body required" \
  bash -c 'printf "" | camp-buzz post'
expect_failure "$output_dir/fail-gate.out" "gate must be" \
  camp-buzz post -m rejected --gate maybe

printf '%s\n' \
  "bind_normalized_http=PASS" \
  "doctor_ready=PASS" \
  "flag_post=PASS" \
  "stdin_post=PASS" \
  "hook_post=PASS" \
  "missing_key_failure=PASS" \
  "invalid_channel_failure=PASS" \
  "empty_stdin_failure=PASS" \
  "invalid_gate_failure=PASS" \
  "disposable_channel=PASS"
