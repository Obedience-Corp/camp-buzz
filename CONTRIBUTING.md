# Contributing

Thank you for contributing to camp-buzz.

## Development

Requirements are Go as declared in `go.mod`, `just`, Git, and Docker. Fork the
repository, create a focused branch, and add tests for changed behavior.

Run the release gate before opening a pull request:

```bash
just release gate
```

Filesystem integration tests run only in the container harness. Never use real
credentials or private Buzz state in tests, fixtures, issues, or pull requests.

## Pull requests

Explain what changed, why it is needed, and how it was verified. Keep public
interfaces backward compatible unless the breaking change is explicit. Update
documentation with behavior changes and accept the
[Code of Conduct](./CODE_OF_CONDUCT.md).

By contributing, you agree that your contribution is licensed under Apache
License 2.0.
