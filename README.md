# camp-buzz

Optional **Buzz** integration plugin for [camp](https://github.com/Obedience-Corp/camp).

This is **not** native Buzz support in camp or fest. It is a standalone
`camp-*` plugin binary. When installed on your `PATH`, camp discovers it as:

```bash
camp buzz …
```

**Design:** campaign workitem WI-ca719b  
`workflow/design/festival-buzz-integration` (Obey campaign).

## What it does

| Command | Purpose |
|---------|---------|
| `camp buzz doctor` | Check buzz CLI, env, bind config (never prints secrets) |
| `camp buzz post` | Post Festival status to a Buzz channel (footer-normalized) |
| `camp buzz bind` | Write `.campaign/integrations/buzz.yaml` (non-secret) |
| `camp buzz show` | Show resolved config (secrets redacted) |
| `camp buzz hook-install` | Print example fest hook YAML |

Posts shell out to the external **`buzz` CLI**. Festival remains source of
truth; Buzz is a projection surface.

## Install (from source)

```bash
just install
# or
go install github.com/Obedience-Corp/camp-buzz/cmd/camp-buzz@latest
```

Optional assets:

```bash
just install-assets   # → ~/.obey/plugins/camp-buzz/
```

## Configure

```bash
export BUZZ_PRIVATE_KEY=…     # secret — never commit
export BUZZ_RELAY_URL=ws://localhost:3000   # or bind file
camp buzz bind --channel <uuid> --relay ws://localhost:3000 --festival CI0009
camp buzz doctor
```

## Fest hooks (optional)

```bash
camp buzz hook-install
# merge into campaign/festival fest hooks config; map events with fest hooks list
```

Hooks must **fail open** — do not block fest advance on post failures.

## Development

```bash
just build
just test
just run doctor
```

## License

Private Obedience Corp repository. Not an open-source release surface yet.
