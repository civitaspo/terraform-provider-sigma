# Releasing

The shared [csm-ci release flow](https://github.com/civitaspo/csm-ci-actions/blob/main/docs/release.md) runs through [civitaspo/csm-ci-server](https://github.com/civitaspo/csm-ci-server). The server configuration selects `build_type: terraform-provider` and `provider_name: sigma`.

## Prepare and merge a release

1. Merge Conventional Commit pull requests through `/approve` and `/merge`. Do not edit `.release-version` or `CHANGELOG.md` on feature PRs.
2. `csm-ci-release-pr` computes the version and full changelog with git-cliff, then requests a signed server update on `release/next`. To select a version explicitly, dispatch that workflow with `version` set to the version without `v`.
3. Review the generated `release-pr: chore(release): vX.Y.Z` PR. It changes only `.release-version` and `CHANGELOG.md`. The provider source, dependencies and packaging configuration stay unchanged.
4. Wait for full CI on the exact release PR head. Comment `/approve`, then `/merge` after policy and approval pass. Release PRs are not automerged.
5. `csm-ci-release` requests publication for that protected merge. The server creates the annotated tag, builds all thirteen platform ZIPs without credentials, and validates the complete package before signing and publication.

The preparation caller sets `update-root-version: false`. `.release-version` supplies the binary version through the existing GoReleaser linker setting. The caller skips preparation on release merge commits so it does not race the publisher for the same version.

## Published assets

Each `vX.Y.Z` release contains exactly seventeen assets:

- Thirteen `terraform-provider-sigma_X.Y.Z_<os>_<arch>.zip` archives: Darwin amd64/arm64; FreeBSD 386/amd64/arm/arm64; Linux 386/amd64/arm/arm64; Windows 386/amd64/arm64.
- `terraform-provider-sigma_X.Y.Z_manifest.json`, preserving the root protocol-6 manifest bytes.
- `terraform-provider-sigma_X.Y.Z_SHA256SUMS`, covering the thirteen ZIPs and the manifest, and its binary detached `.sig`.
- `SHA256SUMS`, covering the other sixteen assets.

Each ZIP contains a root `terraform-provider-sigma_vX.Y.Z` binary, with `.exe` on Windows, and optional root LICENSE, README.md and CHANGELOG.md files. The isolated build reuses the existing GoReleaser archive settings. The trusted publisher validates ZIPs without extracting or executing them, regenerates checksums, and signs with the exact Registry-registered primary RSA key and fixed merge timestamp. It verifies the public signature before attesting every asset and publishing the immutable release. A retry must preserve all asset bytes and the existing immutable release.

## Verify a production release

Check the complete asset set, GitHub asset digests, both checksum files, the binary GPG signature and all seventeen provenance subjects. The public signing fingerprint remains `12D7B26BEB5394D7AA579FB300F0D373EFFA1DC6`. Verify actual new-version discovery in the [Terraform Registry](https://registry.terraform.io/providers/civitaspo/sigma/latest), then run `terraform init` with that exact version and inspect `terraform providers schema -json` in an isolated directory. Registry-side resync and App permission changes belong to the owner.

Run `mise run lint`, `mise run test` and `mise run build` before the release. Keep both handwritten package coverage gates at 80%. Live acceptance tests require separate Sigma credentials and remain a local/manual check.

## Credentials and protection

This client holds only `CSM_CI_CLIENT_APP_PRIVATE_KEY`. Server and Approver App keys, `CSM_CI_TERRAFORM_PROVIDER_GPG_PRIVATE_KEY` and `CSM_CI_TERRAFORM_PROVIDER_GPG_PASSPHRASE` belong only in the production server's `csm-ci` environment. Its `CSM_CI_TERRAFORM_PROVIDER_GPG_FINGERPRINT` variable retains the complete public fingerprint above. Agents never create, read, register or change private keys, tokens or App permissions.

Only the Server App can publish `v*` release tags or perform the controlled squash merge. Immutable release assets cannot be replaced after publication. Preserve unrelated repository protections while transitioning the saved legacy rulesets.

For request workflows and dependency locks, see [securefix.md](securefix.md).
