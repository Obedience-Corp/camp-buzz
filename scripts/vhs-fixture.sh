#!/usr/bin/env bash
# Prepare an isolated fixture for VHS recordings / manual demos.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
root="${CAMP_BUZZ_VHS_ROOT:-/tmp/camp-buzz-vhs-$$}"

rm -rf "$root"
mkdir -p "$root/bin" "$root/home" "$root/campaign"

# camp-buzz binary
(
  cd "$repo_root"
  go build -o "$root/bin/camp-buzz" ./cmd/camp-buzz
)

# fake buzz CLI
cp "$repo_root/scripts/fake-buzz" "$root/bin/buzz"
chmod +x "$root/bin/buzz" "$root/bin/camp-buzz"

# plugin assets
mkdir -p "$root/home/.obey/plugins/camp-buzz"
cp -R "$repo_root/assets/." "$root/home/.obey/plugins/camp-buzz/"

# minimal camp (camp init if available, else hand-roll .campaign)
if command -v camp >/dev/null 2>&1; then
  camp init "$root/campaign" \
    --name camp-buzz-demo \
    --description "VHS fixture for camp-buzz" \
    --mission "Demonstrate optional Buzz projection" \
    --no-register --no-skills >/dev/null 2>&1 || true
fi
mkdir -p "$root/campaign/.campaign/integrations"
# empty bind so bind command can write

# env for demos
cat >"$root/env.sh" <<EOF
export HOME="$root/home"
export PATH="$root/bin:\$PATH"
export CAMP_ROOT="$root/campaign"
export CAMP_BUZZ_ROOT="$root/home/.obey/plugins/camp-buzz"
export BUZZ_PRIVATE_KEY="demo-nsec-not-real"
export BUZZ_FAKE_LOG="$root/fake-buzz.log"
export TERM=xterm-256color
export COLORTERM=truecolor
export PS1=\$'\\033[38;5;208m❯\\033[0m '
cd "$root/campaign"
EOF

echo "$root"
