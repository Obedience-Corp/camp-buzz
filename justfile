#!/usr/bin/env just --justfile
# camp-buzz — optional Buzz integration plugin for camp

set dotenv-load := true

binary_name := "camp-buzz"
bin_dir := "bin"
main_path := "./cmd/camp-buzz"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")

[doc('Cross-platform builds')]
mod xbuild '.justfiles/build.just'

[doc('Testing')]
mod test '.justfiles/test.just'

[doc('Release and versioning')]
mod release '.justfiles/release.just'

[private]
default:
    @echo "camp-buzz — camp plugin (Buzz status projection)"
    @echo ""
    @just --list --unsorted

[no-cd]
build:
    @mkdir -p {{bin_dir}}
    go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

[no-cd]
fmt:
    go fmt ./...

[no-cd]
vet:
    go vet ./...

[no-cd]
lint: fmt vet
    @echo "Lint complete"

[no-cd]
tidy:
    go mod tidy

[no-cd]
clean:
    rm -rf {{bin_dir}} dist out completions

install: build
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p {{gobin}}
    cp {{bin_dir}}/{{binary_name}} {{gobin}}/{{binary_name}}
    if [[ "$(uname)" == "Darwin" ]]; then
        codesign -f -s - {{gobin}}/{{binary_name}} 2>/dev/null || true
    fi
    echo "Installed {{binary_name}} to {{gobin}}/{{binary_name}}"
    echo "camp discovers it as: camp buzz …"

uninstall:
    @rm -f {{gobin}}/{{binary_name}}
    @echo "Removed {{gobin}}/{{binary_name}}"

run *ARGS:
    go run {{main_path}} {{ARGS}}

install-assets:
    #!/usr/bin/env bash
    set -euo pipefail
    dest="${HOME}/.obey/plugins/camp-buzz"
    mkdir -p "$dest"
    cp -R assets/. "$dest/"
    echo "Installed assets to $dest"

# ── VHS demos (requires vhs, ttyd, ffmpeg) ──────────────────────────

[no-cd]
vhs-fixture:
    #!/usr/bin/env bash
    set -euo pipefail
    export CAMP_BUZZ_VHS_ROOT="${CAMP_BUZZ_VHS_ROOT:-/tmp/camp-buzz-vhs}"
    bash scripts/vhs-fixture.sh

[no-cd]
vhs-tour: build
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{justfile_directory()}}"
    mkdir -p docs/demos
    vhs docs/demos/camp-buzz-tour.tape

[no-cd]
vhs-post: build
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{justfile_directory()}}"
    mkdir -p docs/demos
    vhs docs/demos/camp-buzz-post.tape

[no-cd]
vhs: vhs-tour vhs-post
    @echo "Wrote docs/demos/camp-buzz-tour.gif and docs/demos/camp-buzz-post.gif"

[no-cd]
smoke:
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{justfile_directory()}}"
    export CAMP_BUZZ_VHS_ROOT=/tmp/camp-buzz-smoke-$$
    root=$(bash scripts/vhs-fixture.sh)
    # shellcheck disable=SC1091
    source "$root/env.sh"
    camp-buzz version | grep -q camp-buzz
    camp-buzz bind --channel 33333333-3333-4333-8333-333333333333 --relay http://localhost:3000 --festival SMOKE1
    camp-buzz doctor
    camp-buzz show | grep -q SMOKE1
    camp-buzz post -m "smoke status" --task FEST-smoke --gate pass | grep -q posted
    grep -q "smoke status" "$BUZZ_FAKE_LOG"
    grep -q "festival: SMOKE1" "$BUZZ_FAKE_LOG"
    camp-buzz hook-install | grep -q buzz_status
    echo "SMOKE OK ($root)"

# Real buzz CLI + live local relay (requires: built buzz binary, relay on :3000).
# Example:
#   export BUZZ_BIN=/path/to/buzz BUZZ_PRIVATE_KEY=… 
#   just smoke-real
[no-cd]
smoke-real:
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{justfile_directory()}}"
    : "${BUZZ_BIN:?set BUZZ_BIN to the real buzz binary}"
    : "${BUZZ_PRIVATE_KEY:?set BUZZ_PRIVATE_KEY}"
    export BUZZ_RELAY_URL="${BUZZ_RELAY_URL:-http://localhost:3000}"
    if ! curl -sS -m 2 "$BUZZ_RELAY_URL/" >/dev/null; then
      echo "relay not reachable at $BUZZ_RELAY_URL (start: just relay in buzz tree)" >&2
      exit 1
    fi
    root=/tmp/camp-buzz-smoke-real-$$
    rm -rf "$root"
    mkdir -p "$root/bin" "$root/campaign/.campaign" "$root/home"
    go build -o "$root/bin/camp-buzz" ./cmd/camp-buzz
    cp "$BUZZ_BIN" "$root/bin/buzz"
    chmod +x "$root/bin/buzz"
    export PATH="$root/bin:$PATH"
    export CAMP_ROOT="$root/campaign"
    export HOME="$root/home"
    # create a disposable channel on the real relay
    ch_json=$(buzz channels create --name "camp-buzz-smoke-$$" --type stream --visibility open)
    channel=$(echo "$ch_json" | python3 -c 'import sys,json; print(json.load(sys.stdin)["channel_id"])')
    camp-buzz bind --channel "$channel" --relay "$BUZZ_RELAY_URL" --festival REALSMOKE
    camp-buzz doctor
    camp-buzz post -m "real buzz e2e from camp-buzz" --task FEST-real --gate pass
    # confirm message is on the relay
    got=$(buzz messages get --channel "$channel" --limit 3)
    echo "$got" | grep -q "real buzz e2e from camp-buzz"
    echo "$got" | grep -q "festival: REALSMOKE"
    echo "SMOKE-REAL OK channel=$channel"
