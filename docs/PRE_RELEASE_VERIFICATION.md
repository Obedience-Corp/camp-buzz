# Pre-release verification (camp-buzz)

Last run: **2026-08-11** against:

- Real `buzz` CLI (`buzz-cli` release build from Buzz monorepo)
- Live local relay (`just relay` on `http://localhost:3000`)
- Real Nostr keypair (`buzz-admin generate-key`)

## Results

| Gate | Status |
|------|--------|
| `go test ./...` | Pass |
| `just smoke` (fake buzz fixture) | Pass |
| `just smoke-real` (live relay) | Pass |
| Create channel via real `buzz` | Pass |
| `camp-buzz bind` (ws→http normalize) | Pass — file stores `http://…` |
| `camp-buzz doctor` ready | Pass |
| `post -m` with Festival footer | Pass — accepted by relay |
| `post` via stdin | Pass |
| `post --from-hook` default body | Pass |
| `post --no-footer` | Pass — no footer markers |
| `buzz messages get` read-back | Pass — footer fields present |
| Missing channel / missing key | Pass — non-zero exit, clear errors |
| VHS demos (`just vhs`) | Pass (fake buzz; deterministic) |

## Commands to re-run before public open

```bash
# Buzz monorepo
just setup && just relay   # if not already running

export BUZZ_BIN=$PWD/target/release/buzz   # after: cargo build -p buzz-cli --release
export BUZZ_PRIVATE_KEY=…                  # buzz-admin generate-key (secret key)
export BUZZ_RELAY_URL=http://localhost:3000

cd /path/to/camp-buzz
go test ./...
just smoke
just smoke-real
```

## Still out of scope for this gate

- Hosted / production Block relay (needs your membership + real nsec)
- Desktop UI of Buzz (CLI path only)
- Festival lifecycle hooks firing automatically (manual hook config still operator-owned)
- `festival install camp-buzz` until first GitHub release tag + marketplace PR merged

## Ship recommendation

**Plugin core path is good to open** once you:

1. Cut `v0.1.0` release (`just release stable`) so assets exist  
2. Merge marketplace PR (#2)  
3. Optionally flip repo public (or keep private if install auth is handled)

Do **not** claim “works on Block’s hosted relay” until someone runs `smoke-real` against that URL with a real membership key.
