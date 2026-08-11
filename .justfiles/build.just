# Build recipes
binary_name := "camp-buzz"
bin_dir := "bin"
main_path := "./cmd/camp-buzz"

[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

[no-cd]
dev:
    @mkdir -p {{bin_dir}}
    go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

[no-cd]
linux:
    @mkdir -p {{bin_dir}}
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

[no-cd]
linux-arm64:
    @mkdir -p {{bin_dir}}
    GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-arm64 {{main_path}}

[no-cd]
darwin:
    @mkdir -p {{bin_dir}}
    GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

[no-cd]
darwin-arm64:
    @mkdir -p {{bin_dir}}
    GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

[no-cd]
all: linux linux-arm64 darwin darwin-arm64
    @echo "Cross-platform builds complete"
