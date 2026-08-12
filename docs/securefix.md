# Securefix setup

This repository uses [`csm-actions/securefix-action`](https://github.com/csm-actions/securefix-action) and related CSM actions so pull request workflows can request signed commits, approvals, and releases without holding strong credentials.

The shared server repository is [`civitaspo/securefix-server`](https://github.com/civitaspo/securefix-server).

**Canonical client release specification:** [securefix-server docs/client-releases.md](https://github.com/civitaspo/securefix-server/blob/main/docs/client-releases.md)

## GitHub Apps

Install both apps on this repository and on `civitaspo/securefix-server`:

- Client app: `issues: write` (creates request labels on the server repo)
- Server app: `contents: write`, `actions: read`, `pull_requests: write` / `pull_requests: read`, and `workflows: write` as required by Securefix and Release

## Variables and secrets (this repository)

Required for the shared release / approve reusables:

- Secret `SECUREFIX_CLIENT_PRIVATE_KEY` (client GitHub App private key)

Still used by the local `Lint` Securefix autofix path (not hardcoded in that workflow yet):

- Variable `SECUREFIX_CLIENT_APP_ID`
- Variable `SECUREFIX_SERVER_REPOSITORY` (value: `securefix-server` — repository name only; `securefix-action` expects this format)

Strong secrets (GPG keys, machine-user PAT, server app private key) live only in `civitaspo/securefix-server`.

## Flows

### Lint autofix (Securefix)

The `Lint` workflow runs fixers (`pinact`, `disable-checkout-persist-credentials`, `gofmt`, `go mod tidy`, `tfplugindocs`). When a pull request needs fixes, it requests a Securefix commit. The server workflow accepts client workflows named `Lint` and `Release PR`.

### Auto-approval

The `Approve Request` thin wrapper calls `reusable-approve-request.yml` on `securefix-server`. Trusted authors / committers are defined on the server (see client-releases.md). When validation passes, the server approves with the machine-user PAT (`civitaspo-bot`).

### Release

1. The `Release PR` reusable runs git-cliff (`cliff.toml` + mise `aqua:orhun/git-cliff`) and asks Securefix to open or update `release/next`.
2. After that PR is squash-merged, the `Release Tag` reusable creates an annotated tag and a `release-request-*` label on `civitaspo/securefix-server`.
3. The server `Release` workflow (`publish: goreleaser` for this repo in `release-clients.yaml`) runs GoReleaser with the provider GPG key from the server `main` environment and publishes the GitHub Release.

See [releasing.md](releasing.md) and the [canonical client-releases doc](https://github.com/civitaspo/securefix-server/blob/main/docs/client-releases.md).
