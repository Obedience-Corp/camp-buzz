# Windows artifact candidates

The existing GoReleaser build remains the source of truth. Its Windows amd64
build reuses the common configuration and runs only for snapshots with
WINDOWS_CANDIDATE=true. Normal tag releases cannot publish that build, even if the
environment flag is set. Existing macOS/Linux public outputs and payloads remain
unchanged; no unsigned Windows support is advertised by this change.

`just release windows-candidate <X.Y.Z>` creates an unpublished five-platform
snapshot. The existing artifact verifier has an explicit --windows-candidate mode
that checks camp-buzz.exe, PE amd64 architecture, all runtime templates/completions,
and checksums. Default production verification still requires exactly the existing
four platform archives. Missing/bad PE or binary naming is rejected in container
integration tests.

Run `just release gate` for the established full isolated gates and
`just release candidate-gate` for both default/four-platform and opt-in/five-platform
snapshots plus native Linux extracted execution. Git history is mounted using the
common repository directory, which works in dedicated worktrees. Fixture tests
never inherit release-only GIT_DIR variables.

The manual Windows candidate workflow executes the same archive bytes on disposable
Windows/macOS/Linux runners. It does not publish or claim trusted Windows signatures.
008 must establish actual Windows operational/public release proof and document the tested signing/trust state before enabling Windows publication or app support. Do not advertise Authenticode or Smart App Control acceptance without that evidence.
No signing credential or publication authority is inferred here.

Public unsigned macOS launch is accepted with documented per-app Open Anyway. Apple Developer ID/notarization is optional post-launch and cannot gate the existing macOS/Linux public release. Windows candidates do not establish Authenticode or Smart App Control trust. Checksums, signed marketplace metadata and app updater signatures remain required.
