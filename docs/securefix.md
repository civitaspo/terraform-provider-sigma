# Securefix setup

This repository uses [`csm-actions/securefix-action`](https://github.com/csm-actions/securefix-action) and related CSM actions so pull request workflows can request signed commits, approvals, and releases without holding strong credentials.

The shared server repository is [`civitaspo/securefix-server`](https://github.com/civitaspo/securefix-server).

## GitHub Apps

Install both apps on this repository and on `civitaspo/securefix-server`:

- Client app: `issues: write` (creates request labels on the server repo)
- Server app: `contents: write`, `actions: read`, `pull_requests: write`, and `workflows: write`

## Variables and secrets (this repository)

- Variable `SECUREFIX_CLIENT_APP_ID`
- Secret `SECUREFIX_CLIENT_PRIVATE_KEY`
- Variable `SECUREFIX_SERVER_REPOSITORY` (value: `securefix-server` — repository name only; `securefix-action` expects this format)

Strong secrets (GPG keys, machine-user PAT, server app private key) live only in `civitaspo/securefix-server`.

## Flows

### Lint autofix (Securefix)

The `Lint` workflow runs fixers (`pinact`, `disable-checkout-persist-credentials`, `gofmt`, `go mod tidy`, `tfplugindocs`). When a pull request needs fixes, it requests a Securefix commit. The server workflow accepts client workflows named `Lint` and `Release PR`.

### Auto-approval

The `Approve Request` workflow asks the server to approve pull requests authored by `civitaspo` or `renovate[bot]` (and on `/approve` comments from `civitaspo`). The server validates that all commits are signed and that committers are in an allowlist, then approves with a machine-user fine-grained PAT.

### Release

1. The `Release PR` workflow asks Securefix to open or update `release/next` with changelog and version metadata.
2. After that PR is squash-merged, the `Release Tag` workflow creates an annotated tag and requests a server-side release.
3. The server `main` environment runs GoReleaser with the provider GPG key and publishes the GitHub Release.

See [releasing.md](releasing.md) once the release pipeline lands.
