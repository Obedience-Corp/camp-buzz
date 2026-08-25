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

export PATH="${BUZZ_BIN_DIR}:${PATH}"
export HOME="${CAMP_BUZZ_SMOKE_HOME:-/tmp/camp-buzz-smoke-home}"
export CAMP_ROOT="${CAMP_BUZZ_SMOKE_ROOT:-/tmp/camp-buzz-smoke-campaign}"
mkdir -p "$HOME" "$CAMP_ROOT/.campaign"

BUZZ_PRIVATE_KEY="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"
export BUZZ_PRIVATE_KEY
trap 'unset BUZZ_PRIVATE_KEY' EXIT

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
  --relay ws://localhost:3030 \
  --festival CB0001 \
  --festival-path festivals/active/camp-buzz-first-public-release-CB0001 \
  >/tmp/camp-buzz-bind.out
grep -q "wrote" /tmp/camp-buzz-bind.out
grep -q "relay_url: http://localhost:3030" \
  "$CAMP_ROOT/.campaign/integrations/buzz.yaml"

doctor="$(camp-buzz doctor)"
printf '%s\n' "$doctor" | grep -q "status: ready"
printf '%s\n' "$doctor" | grep -q "BUZZ_PRIVATE_KEY: set"

camp-buzz post \
  -m "CB0001 flag path" \
  --task CB0001:P001.S03.T02 \
  --gate pass \
  >/tmp/camp-buzz-flag.out
printf "CB0001 stdin path\n" | camp-buzz post \
  --task CB0001:P001.S03.T02-stdin \
  --gate pending \
  >/tmp/camp-buzz-stdin.out

hook_yaml="$(camp-buzz hook-install)"
printf '%s\n' "$hook_yaml" | grep -q "command: camp buzz post --from-hook"
printf '%s\n' "$hook_yaml" | grep -q "fail: open"
camp-buzz post \
  --from-hook \
  --task CB0001:P001.S03.T02-hook \
  --gate n/a \
  >/tmp/camp-buzz-hook.out

buzz messages get --channel "$channel" --limit 10 >/tmp/camp-buzz-messages.json
python3 - /tmp/camp-buzz-messages.json <<'PY'
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

expect_failure /tmp/camp-buzz-fail-key.out "BUZZ_PRIVATE_KEY is not set" \
  env -u BUZZ_PRIVATE_KEY camp-buzz post -m rejected
expect_failure /tmp/camp-buzz-fail-channel.out "channel id must be a UUID" \
  camp-buzz post -m rejected --channel not-a-uuid
expect_failure /tmp/camp-buzz-fail-empty.out "message body required" \
  bash -c 'printf "" | camp-buzz post'
expect_failure /tmp/camp-buzz-fail-gate.out "gate must be" \
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
