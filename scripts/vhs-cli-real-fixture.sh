#!/usr/bin/env bash
# Stage an isolated env for real-CLI VHS recording (live relay + real buzz).
# Prints the fixture root; writes $root/env.sh for the tape to source.
set -euo pipefail

: "${BUZZ_BIN:?set BUZZ_BIN to the real buzz binary}"
: "${BUZZ_PRIVATE_KEY:?set BUZZ_PRIVATE_KEY}"
export BUZZ_RELAY_URL="${BUZZ_RELAY_URL:-http://localhost:3000}"
export BUZZ_AUTH_TAG="${BUZZ_AUTH_TAG:-}"

if ! curl -sS -m 2 "${BUZZ_RELAY_URL}/" >/dev/null; then
  echo "relay not reachable at $BUZZ_RELAY_URL" >&2
  exit 1
fi

ROOT="${CAMP_BUZZ_CLI_REAL_ROOT:-/tmp/camp-buzz-cli-real-demo}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
rm -rf "$ROOT"
mkdir -p "$ROOT/bin" "$ROOT/campaign/.campaign" "$ROOT/home"

if [[ ! -x "$REPO_ROOT/bin/camp-buzz" ]]; then
  (cd "$REPO_ROOT" && go build -o bin/camp-buzz ./cmd/camp-buzz)
fi
cp "$REPO_ROOT/bin/camp-buzz" "$ROOT/bin/camp-buzz"
cp "$BUZZ_BIN" "$ROOT/bin/buzz"
chmod +x "$ROOT/bin/"*

export PATH="$ROOT/bin:$PATH"
export HOME="$ROOT/home"
export CAMP_ROOT="$ROOT/campaign"

ch_json=$(buzz channels create --name "camp-buzz-cli-demo" --type stream --visibility open)
channel=$(echo "$ch_json" | python3 -c 'import sys,json; print(json.load(sys.stdin)["channel_id"])')
camp-buzz bind --channel "$channel" --relay "$BUZZ_RELAY_URL" --festival DEMO-UI >/dev/null

# env.sh is sourced inside the VHS session (secrets never typed on camera).
cat >"$ROOT/env.sh" <<EOF
export PATH="$ROOT/bin:\$PATH"
export HOME="$ROOT/home"
export CAMP_ROOT="$ROOT/campaign"
export BUZZ_RELAY_URL="$BUZZ_RELAY_URL"
export BUZZ_PRIVATE_KEY="$BUZZ_PRIVATE_KEY"
export BUZZ_AUTH_TAG="${BUZZ_AUTH_TAG:-}"
export CAMP_BUZZ_DEMO_CHANNEL="$channel"
export PS1=\$'\\033[38;5;208m❯\\033[0m '
EOF

echo "$ROOT"
