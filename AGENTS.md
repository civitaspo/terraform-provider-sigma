# Repository Guidelines

## Project Scope

This repository contains the Terraform provider for [Sigma Computing](https://www.sigmacomputing.com/), published as `registry.terraform.io/civitaspo/sigma`.

## Contributor Expectations

- Write commits, pull request titles/bodies, documentation, comments, and user-facing messages in **English only**.
- Use Conventional Commits for pull request titles (`feat`, `fix`, `docs`, `refactor`, `test`, `ci`, `build`, `chore`, `perf`, `revert`; use `!` for breaking changes).
- Never push directly to `main`. Open a pull request and squash-merge after required checks pass.
- Keep changes small, reviewable, and focused on one meaningful unit of work.
- Sign commits (SSH signing is configured for maintainers and coding agents committing as `civitaspo`).
- Do not store strong credentials in this repository. GPG keys, machine-user PATs, and `contents: write` app keys live only in `civitaspo/securefix-server`.

## Tooling

Install pinned tools with mise:

```bash
mise install --locked
```

Before opening a pull request, run:

```bash
mise run lint
mise run test
```

Useful tasks:

- `mise run build` — `go build ./...`
- `mise run docs` — regenerate provider docs with `tfplugindocs` (never hand-edit files under `docs/` except `docs/securefix.md` and `docs/releasing.md`)
- `mise run openapi-generate` — generate the Sigma REST client from `specs/sigma-rest-api.openapi.json` (never hand-edit `internal/sigma/openapi/generated.go`)
- `mise run openapi-check` — regenerate from the vendored snapshot and fail on generated drift
- `mise run openapi-update` — fetch a new live OpenAPI snapshot, normalize it, and regenerate (manual only; never run in CI)

## GitHub Actions

- Pin public GitHub Actions to immutable SHAs.
- Use `persist-credentials: false` with `actions/checkout` unless a workflow explicitly needs push credentials.
- Keep workflow permissions least-privilege.
- Securefix applies machine fixes (pinact, gofmt, tidy, docs drift) via `civitaspo/securefix-server`.
- Approvals for trusted authors are requested through `csm-actions/approve-pr-action`.

See [CONTRIBUTING.md](CONTRIBUTING.md), [docs/securefix.md](docs/securefix.md), and [docs/releasing.md](docs/releasing.md).

## Verification

- Unit tests use a mock Sigma HTTP server and must pass with `go test -race -cover ./...`.
- Acceptance tests (`TF_ACC=1`) require real Sigma credentials and are optional locally.
- Do not invent API field names; verify against `https://help.sigmacomputing.com/reference/<slug>.md`.
