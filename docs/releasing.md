# Releasing

This repository uses the shared Securefix client release flow hosted in [`civitaspo/securefix-server`](https://github.com/civitaspo/securefix-server).

**Canonical specification:** [securefix-server docs/client-releases.md](https://github.com/civitaspo/securefix-server/blob/main/docs/client-releases.md)

## Overview

1. Commits land on `main` via squash-merged pull requests.
2. **Release PR** (reusable on securefix-server) computes the next version and changelog, then opens or updates `release/next` via Securefix.
3. **Release PR Sync** keeps the open `release/next` PR title/body aligned with `.release-version`.
4. A human squash-merges `chore(release): vX.Y.Z`.
5. **Release Tag** creates annotated tag `vX.Y.Z` and creates a `release-request-*` label on `civitaspo/securefix-server`.
6. The server **Release** workflow (`publish: goreleaser` in the allowlist) checks out the tag, runs GoReleaser with the GPG key from the server `main` environment, and publishes the GitHub Release.

Local workflows under `.github/workflows/` are thin wrappers that `uses:` the securefix-server reusables at a pinned commit SHA.

## Repository release protections

- **Tag protection** (`Protect tags` ruleset): active — blocks force-pushes and deletion of `v*` tags outside the allowed release path.
- **Immutable releases**: enabled — after a release is published, its assets and associated Git tag cannot be modified or deleted.

GoReleaser is compatible with immutable releases: it creates the GitHub Release as a draft, uploads all artifacts, then publishes once. Do not enable workflows that attach or replace assets on an already-published release.

## Version bump rules (while major is 0)

Configured in `cliff.toml` (`[bump]`) and applied by git-cliff on the shared Release PR reusable:

- Conventional Commit breaking change (`type!:` or `BREAKING CHANGE`) → minor
- `feat:` → minor
- everything else releasable → patch
- Only `chore(release):` commits since the last tag → nothing to release

The Release PR workflow floors the base version on both the latest `v*` tag and `.release-version` on `main`.

git-cliff regenerates the full changelog from git history. **Do not edit `CHANGELOG.md` on feature PRs** — Lint fails if a non-`release/next` PR touches that file. Use Conventional Commit subjects; the Release PR is the only writer.

## Local helpers

```bash
mise install --locked
mise exec -- git cliff --bumped-version
mise exec -- git cliff --tag vX.Y.Z --output CHANGELOG.md
goreleaser check
goreleaser build --snapshot --clean --single-target
```

## First registry publication

After `v0.1.0` exists with signed assets:

1. Sign in to https://registry.terraform.io with GitHub.
2. Settings → GPG Keys → add the provider release public key under namespace `civitaspo`.
3. Publish → Provider → select `civitaspo/terraform-provider-sigma`.

## v0.2.0 release criteria

Do not tag `v0.2.0` until all of the following are true:

- `mise run lint`, `mise run test`, and `mise run build` pass on `main`.
- Handwritten `internal/provider` and `internal/sigma` statement coverage are each at least 80% (generated OpenAPI code is excluded).
- Contract tests use the request recorder (unexpected method/path, query mismatch, JSON mismatch, unconsumed requests).
- Live acceptance (`TF_ACC=1`) remains local/manual and is not part of PR CI.
- git-cliff / the release PR records breaking v0.2 resource and attribute removals; do not hand-edit `CHANGELOG.md` on feature PRs.

## Credentials

Repository secret `SECUREFIX_CLIENT_PRIVATE_KEY` only. See [securefix.md](securefix.md) and the [canonical client-releases doc](https://github.com/civitaspo/securefix-server/blob/main/docs/client-releases.md).
