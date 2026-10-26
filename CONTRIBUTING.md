# Contributing

Thanks for contributing to `terraform-provider-sigma`.

## Development setup

Install pinned tools with [mise](https://mise.jdx.dev/):

```bash
mise install --locked
```

Useful tasks:

| Task | Command |
|------|---------|
| Lint | `mise run lint` |
| Unit tests | `mise run test` |
| Build | `mise run build` |
| Regenerate docs | `mise run docs` |
| Generate OpenAPI client from the vendored snapshot | `mise run openapi-generate` |
| Check generated OpenAPI client for drift | `mise run openapi-check` |
| Fetch the live OpenAPI snapshot and regenerate (manual; never in CI) | `mise run openapi-update` |

Never hand-edit generated files under `docs/` except `docs/securefix.md` and `docs/releasing.md`. Never hand-edit `internal/sigma/openapi/generated.go` or `specs/sigma-rest-api.openapi.json`; apply generator corrections only through `internal/sigma/openapi/overlay.yaml` and refresh the snapshot with `mise run openapi-update`.

## Pull requests

- Write commits, PR titles/bodies, documentation, and comments in **English only**.
- Use Conventional Commits for PR titles (`feat`, `fix`, `docs`, `refactor`, `test`, `ci`, `build`, `chore`, `perf`, `revert`; use `!` for breaking changes).
- Never push directly to `main`. Open a PR and squash-merge after required checks pass.
- Keep changes small, reviewable, and focused on one meaningful unit of work.
- Sign commits (SSH signing is configured for maintainers and coding agents committing as `civitaspo`).
- Do **not** edit `CHANGELOG.md` on feature PRs. git-cliff regenerates it on the `release/next` Release PR from Conventional Commit subjects (see [docs/releasing.md](docs/releasing.md)).

Before opening a PR:

```bash
mise run lint
mise run test
```

## Testing

### Unit tests

Unit tests use a mock Sigma HTTP server and must pass with:

```bash
go test -race -cover ./...
# or
mise run test
```

### Acceptance tests

Acceptance tests are gated behind `TF_ACC=1` and require real Sigma credentials:

```bash
export TF_ACC=1
export SIGMA_BASE_URL="https://aws-api.sigmacomputing.com"   # or your cloud
export SIGMA_CLIENT_ID="..."
export SIGMA_CLIENT_SECRET="..."
go test ./internal/provider/ -run TestAcc -count=1 -timeout 30m
```

Many acceptance stubs currently skip until dedicated fixtures exist. Prefer mock-server unit coverage for CRUD when adding resources.

Do not invent API field names; verify against `https://help.sigmacomputing.com/reference/<slug>.md`.

## Release flow (summary)

1. Merges to `main` trigger `csm-ci-release-pr`, which prepares `.release-version` and `CHANGELOG.md` on `release/next` through the server.
2. The owner requests `/approve` and `/merge` after full CI passes on the current release PR head.
3. The server builds all thirteen archives without credentials, validates and GPG-signs them in the trusted publisher, and publishes an immutable release with checksums and provenance for the Terraform Registry.

See [docs/releasing.md](docs/releasing.md) and [docs/securefix.md](docs/securefix.md) for details.

## Security

Do not store strong credentials in this repository. Report vulnerabilities via GitHub private vulnerability reporting ([SECURITY.md](SECURITY.md)).
