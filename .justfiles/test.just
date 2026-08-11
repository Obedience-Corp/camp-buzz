# Test recipes

[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

[no-cd]
all:
    go test ./...

[no-cd]
unit:
    go test ./...

[no-cd]
race:
    go test -race ./...

[no-cd]
short:
    go test -short ./...

[no-cd]
coverage:
    @mkdir -p out
    go test -coverprofile=out/coverage.out ./...
    go tool cover -html=out/coverage.out -o out/coverage.html
    @echo "Coverage report: out/coverage.html"
