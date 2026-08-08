# Releasing

This repository uses a tagpr-equivalent release flow built on CSM actions and [git-cliff](https://git-cliff.org/).

## Overview

1. Commits land on `main` via squash-merged pull requests.
2. The **Release PR** workflow runs git-cliff to bump the version and regenerate `CHANGELOG.md`, then asks `civitaspo/securefix-server` to open or update `release/next`.
3. The **Release PR Sync** workflow updates the open `release/next` pull request title and body from `.release-version` (securefix creates the PR once and does not refresh metadata on later pushes).
4. A human squash-merges `chore(release): vX.Y.Z`.
5. The **Release Tag** workflow creates an annotated tag `vX.Y.Z` and requests a server-side release.
6. The securefix-server **Release Terraform Provider** workflow checks out the tag, runs GoReleaser with the GPG key from the `main` environment, and publishes the GitHub Release.

## Repository release protections

- **Tag protection** (`Protect tags` ruleset): active — blocks force-pushes and deletion of `v*` tags outside the allowed release path.
- **Immutable releases**: enabled — after a release is published, its assets and associated Git tag cannot be modified or deleted.

GoReleaser is compatible with immutable releases: it creates the GitHub Release as a draft, uploads all artifacts, then publishes once. The SecureFix release workflow does not re-upload or mutate assets after publish. Do not enable workflows that attach or replace assets on an already-published release.

## Version bump rules (while major is 0)

Configured in [`cliff.toml`](../cliff.toml) (`[bump]`) and applied by git-cliff:

- Conventional Commit breaking change (`type!:` or `BREAKING CHANGE`) → minor
- `feat:` → minor
- everything else releasable → patch
- Only `chore(release):` commits since the last tag → nothing to release

The Release PR workflow floors the base version on both the latest `v*` tag and `.release-version` on `main`. If the floor tag is not present yet (release merged, Release Tag still running), it creates a **local** tag at the matching `chore(release):` commit so git-cliff does not re-propose the version that just shipped. That local tag is never pushed.

## Local preview

```bash
mise install --locked
mise exec -- git cliff --bumped-version
mise exec -- git cliff --tag vX.Y.Z --output CHANGELOG.md
goreleaser check
goreleaser build --snapshot --clean --single-target
```

git-cliff regenerates the full changelog from git history. **Do not edit `CHANGELOG.md` on feature PRs** — Lint fails if a non-`release/next` PR touches that file. Use Conventional Commit subjects; the Release PR is the only writer. Hand-edited `## Unreleased` notes cause merge conflicts with the long-lived `release/next` branch.

## Server request format

The Release Tag workflow creates a label on `civitaspo/securefix-server` whose description is:

```text
civitaspo/terraform-provider-sigma/<run_id>/vX.Y.Z/<merge-commit-sha>
```

If that string would exceed GitHub's 100-character label description limit, the merge commit SHA is omitted and the server resolves it from the merged `release/next` pull request.

The merge commit SHA is preferred because `Release Tag` on a merged `release/next` PR tags the squash-merge commit on `main`, while the workflow run's `head_sha` is the PR head. The server accepts in-progress `Release Tag` runs and publishes signed artifacts for that tag.

## First registry publication

After `v0.1.0` exists with signed assets:

1. Sign in to https://registry.terraform.io with GitHub.
2. Settings → GPG Keys → add the provider release public key under namespace `civitaspo`.
3. Publish → Provider → select `civitaspo/terraform-provider-sigma`.

See [securefix.md](securefix.md) for client/server credential layout.
