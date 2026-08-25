# Release recipes

[private]
[no-cd]
default:
    @go run ./internal/releaseutil help

goreleaser := "go run github.com/goreleaser/goreleaser/v2@v2.15.2"
releaseutil := "go run ./internal/releaseutil"

[no-cd]
gate:
    bash scripts/release-gate-container.sh

[no-cd]
check:
    {{goreleaser}} check

[no-cd]
snapshot:
    {{goreleaser}} release --snapshot --clean

[no-cd]
package-check TAG:
    {{goreleaser}} release --clean --skip=publish
    {{releaseutil}} artifacts --dir dist --tag {{TAG}}

[no-cd]
version:
    @go run ./cmd/camp-buzz version

[no-cd]
current:
    @{{releaseutil}} current

[no-cd]
stable level="patch":
    {{releaseutil}} stable --level {{level}}

[no-cd]
tag VERSION:
    {{releaseutil}} tag --version {{VERSION}}
