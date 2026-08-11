#!/usr/bin/env just --justfile
# camp-buzz — optional Buzz integration plugin for camp

set dotenv-load := true

binary_name := "camp-buzz"
bin_dir := "bin"
main_path := "./cmd/camp-buzz"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")

[private]
default:
    @echo "camp-buzz — camp plugin (Buzz status projection)"
    @echo ""
    @just --list --unsorted

build:
    @mkdir -p {{bin_dir}}
    go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

fmt:
    go fmt ./...

vet:
    go vet ./...

lint: fmt vet
    @echo "Lint complete"

test:
    go test ./...

tidy:
    go mod tidy

clean:
    rm -rf {{bin_dir}} dist out

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
