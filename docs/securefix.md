# Pull request automation

This repository uses [csm-ci-actions](https://github.com/civitaspo/csm-ci-actions) with [civitaspo/csm-ci-server](https://github.com/civitaspo/csm-ci-server).

## Client configuration

The client uses variables `CSM_CI_CLIENT_APP_CLIENT_ID` and `CSM_CI_SERVER_REPOSITORY` (`civitaspo/csm-ci-server`), and secret `CSM_CI_CLIENT_APP_PRIVATE_KEY`. Strong credentials stay in the server repository's `csm-ci` environment. Credential and App permission changes belong to the owner.

## Approval and merge

The managed callers request policy checks and approval for signed commits by trusted committers. Sensitive changes, including workflows and dependency manifests, require the owner to comment `/approve` on the current head, except for Renovate PRs covered by the server's explicit sensitive-path exemption. That exemption removes the owner-comment requirement; the required server approval and checks still apply. A new commit dismisses the old approval. `/recheck` refreshes the policy and approval decision.

After the required `status-check`, `csm-ci/policy`, and approval pass, the owner comments `/merge`. A configured trusted Bot may request `/merge` only on its own PR when every commit is its own; the same current-head approval, policy and CI requirements apply. Renovate has an explicit sensitive-path exemption in the server policy. The Server App performs the squash merge. Repository settings disable other merge methods and automatic merging. Common caller workflows and the PR template are managed by the server.

## Workflow autofix

The Autofix job checks out the exact PR head, runs the existing workflow fixers, and uses `actions/client/securefix` to package changes. The server validates the workflow run, signed head, artifact, and configured write policy before committing fixes. Workflow changes are explicitly allowed by the shared server policy. No changes means no request.

Renovate owns GitHub Actions and mise version updates. Dependabot security updates remain eligible for the same trusted-commit policy. Dependency PRs still pass the required checks and use the server merge flow.

For releases, see [releasing.md](releasing.md).

## Workflow dependency lockfiles

`.github/workflows/actions.lock` records direct pins and transitive action dependencies. Full SHA pins, `pinact`, and the SHA-pinning policy remain required. After changing dependencies, regenerate the lockfile with `mise run actions-lock` and verify it with `mise run actions-lock-verify` using the ordinary GitHub login. Verification never writes files. See the [development commands](../README.md#workflow-dependencies).

The shared actions repository also commits its workflow lockfile. This client lockfile does not replace that coverage.
