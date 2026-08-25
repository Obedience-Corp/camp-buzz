# Release verification (camp-buzz)

Last run: **2026-08-25**

```text
snapshot proof commit: 402fb44adfa881468d61369d74d8a9e65df3be89
reviewed PR head: fab7a38e051c4564cdc4dec237e1db4f1f6e9224
merged public baseline: b97ca786120fd3eb0cc290ff93eb8c0b79250030
block/buzz origin/main: 113a33b7e49b7173ee1767c49ef2f49c63803034
Go: 1.25.9
GoReleaser: 2.15.2
```

The changes after the snapshot proof and through the reviewed PR head hardened
the reusable smoke harness and public documentation; they did not change the Go
runtime or packaged artifact surface. GitHub merged that reviewed head as the
public baseline shown above.

## Proven scope

- Exact-current Buzz CLI built from source in an isolated Rust 1.95.0
  container.
- Live, disposable local Buzz relay with isolated Postgres, Redis, and MinIO.
- Random one-time private key and 300-second channel generated inside the
  smoke container; neither value was printed or persisted.
- `camp-buzz` built in the pinned Go container from a read-only source mount.
- Clean GoReleaser snapshot built from the exact committed candidate.

This proves the current local CLI/relay path. It does **not** claim
compatibility with Block's hosted or production relay.

The verified public release is
[`v0.1.2`](https://github.com/Obedience-Corp/camp-buzz/releases/tag/v0.1.2).
The earlier tags remain immutable. `v0.1.2` is the fix-forward release whose
binary reports its published version in both release archives and
`go install ...@latest` builds.

## Results

| Gate | Status |
| --- | --- |
| Complete container release gate | Pass |
| Actionlint, Go fmt/vet/test/race/build, Staticcheck | Pass |
| govulncheck | Pass — zero called vulnerabilities |
| Gitleaks working tree and full history | Pass |
| Current Buzz command/env/exit contract audit | Pass |
| Create disposable channel via real current `buzz` | Pass |
| `camp-buzz bind` (`ws` → `http`) | Pass |
| `camp-buzz doctor` ready | Pass |
| Flag, stdin, and hook-path posts | Pass |
| Exact message/footer read-back via real `buzz` | Pass — 3/3 |
| Missing key, bad channel, empty stdin, invalid gate | Pass — clear nonzero errors |
| Snapshot archive/checksum matrix | Pass — 4/4 |
| Packaged docs, assets, completions, license/notice | Pass — every archive |
| Binary OS/architecture metadata | Pass — 4/4 |
| Native snapshot binary `version` and `--help` | Pass |
| Public `v0.1.2` archive/checksum matrix | Pass — 4/4 |
| Anonymous Festival `v0.3.2` marketplace install | Pass |
| Installer-managed executable and runtime templates | Pass |
| `camp buzz version`, isolated doctor/post, Fest hook resolution | Pass |

All disposable relay containers, networks, data/build/cache volumes, generated
key material, channel state, and snapshot artifacts were removed after proof.

## Reproduce without private credentials

1. Fetch `block/buzz` and verify the recorded commit:

   ```bash
   git fetch origin main
   git rev-parse origin/main
   # 113a33b7e49b7173ee1767c49ef2f49c63803034
   ```

2. In disposable containers, build `buzz-cli`, `buzz-relay`, and `buzz-admin`
   from that commit and start the upstream isolated harness on its alternate
   local ports (relay `3030`, health `8088`). Do not use the shared `:3000`
   development relay.

3. Build `camp-buzz` from the candidate in the pinned Go container, then run
   the guarded smoke runner inside the relay-connected container:

   ```bash
   CAMP_BUZZ_CONTAINER_SMOKE=1 \
   BUZZ_BIN_DIR=/out \
   BUZZ_RELAY_URL=http://localhost:3030 \
   bash scripts/smoke-current-buzz.sh
   ```

   The runner generates its own key and channel. It must not be given a real
   key. It refuses to run unless `CAMP_BUZZ_CONTAINER_SMOKE=1` is set.

4. Run the complete release gate:

   ```bash
   just release gate
   ```

## Still out of scope

- Block hosted/production relay membership and authorization.
- Buzz Desktop UI; this proof covers the real CLI path.
- Hosting or operating a relay for users.
- Any requirement that core camp, fest, or Festival users install this plugin.

## Public distribution result

The repository is public, the protected `main` and release workflows passed,
and `v0.1.2` provides the expected four archives plus checksums. A fresh
credential-free Linux/arm64 container installed the plugin from the official
marketplace through Festival `v0.3.2` / festival-installer `v0.1.1`. The
receipt owned the managed binary and both templates under
`~/.obey/plugins/camp-buzz/`; `camp buzz version`, `doctor`, `post`, and
`fest hooks list` all passed using only those installed locations.

Do not claim hosted-relay support until an authorized operator runs a separate
membership-gated smoke and records sanitized evidence.
