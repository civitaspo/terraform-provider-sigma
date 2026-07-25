# Releasing

This repository uses a tagpr-equivalent release flow built on CSM actions.

## Overview

1. Commits land on `main` via squash-merged pull requests.
2. The **Release PR** workflow computes the next version and changelog, then asks `civitaspo/securefix-server` to open or update `release/next`.
3. A human squash-merges `chore(release): vX.Y.Z`.
4. The **Release Tag** workflow creates an annotated tag `vX.Y.Z` and requests a server-side release.
5. The securefix-server **Release Terraform Provider** workflow checks out the tag, runs GoReleaser with the GPG key from the `main` environment, and publishes the GitHub Release.

## Version bump rules (while major is 0)

Scripts live under `scripts/release/`.

- Conventional Commit breaking change (`type!:` or `BREAKING CHANGE`) → minor
- `feat:` → minor
- everything else releasable → patch
- Only `chore(release):` commits since the last tag → nothing to release

The Release PR workflow floors the base version on both the latest `v*` tag and `.release-version` on `main`, so a push that lands after a release merge but before Release Tag finishes cannot re-propose the version that just shipped.

## Local helpers

```bash
scripts/release/next-version.sh
scripts/release/changelog.sh 0.1.0
goreleaser check
goreleaser build --snapshot --clean --single-target
```

## Server request format

The Release Tag workflow creates a label on `civitaspo/securefix-server` whose description is:

```text
civitaspo/terraform-provider-sigma/<run_id>/vX.Y.Z
```

The server accepts in-progress `Release Tag` runs and publishes signed artifacts for that tag.

## First registry publication

After `v0.1.0` exists with signed assets:

1. Sign in to https://registry.terraform.io with GitHub.
2. Settings → GPG Keys → add the provider release public key under namespace `civitaspo`.
3. Publish → Provider → select `civitaspo/terraform-provider-sigma`.

See [securefix.md](securefix.md) for client/server credential layout.
