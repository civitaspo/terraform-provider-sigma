# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.2.1] - 2026-08-19


### Bug Fixes

- update module github.com/oapi-codegen/runtime to v1.7.0 (#99)
- keep configured state and finish live catalog pagination (#96)


### Maintenance

- update dependency jdx/mise to v2026.8.9 (#103)
- update dependency aqua:hashicorp/terraform to v1.15.9 (#101)
- update dependency jdx/mise to v2026.8.8 (#100)
- update dependency jdx/mise to v2026.8.6 (#97)

## [0.2.0] - 2026-08-14


### Bug Fixes

- isolate Beta APIs and replace aggregate attachments with singular relations (#91)
- expose documented data-source filters and split connection path lookup (#90)
- require complete connection PUTs and paired write-only credentials (#89)
- correct dedicated grants, schedules, public embeds, and translations (#88)
- make identity lifecycle and Framework conversions fail closed (#85)
- update module github.com/oapi-codegen/runtime to v1.6.0 (#83)
- install git-cliff for shared Release PR reusable (#66)
- pass merge commit SHA in release request label (#46)


### Documentation

- complete v0.2 out-of-scope declarations (#95)
- document coverage matrix and out-of-scope document content (#78)
- fix broken CI badge after workflow rename (#71)
- point releasing guide at shared client-releases spec (#64)


### Features

- add materialization schedule and account type permission data sources (#94)
- add singular lookup data sources (#93)
- remove overlapping public resources and add sigma_folder (#84)
- wrap generated Sigma REST client with hardened transport (#82)
- add sigma_user_attribute_tenant_assignment (#77)
- add restore and computed use_oauth to sigma_connection (#76)
- add workbook source copy and restore on sigma_file (#75)
- add member create options and destroy-time transfer controls (#74)
- add create_team_folder and workspace_id to sigma_team (#73)


### Maintenance

- add contract fixtures, coverage gates, and v0.2 release criteria (#92)
- update dependency go to v1.26.6 (#87)
- update csm-actions/securefix-action action to v0.6.1 (#86)
- add reproducible Sigma REST OpenAPI client generation (#79)
- update jdx/mise-action action to v4.2.5 (#80)
- update dependency jdx/mise to v2026.8.5 (#72)
- migrate Renovate config (#63)
- harden reusable workflow calls for status-check (#70)
- update dependency jdx/mise to v2026.8.4 (#69)
- grant nested reusable workflow permissions from callers (#68)
- collapse PR checks into status-check gate (#67)
- bump securefix-server reusables for job summary links (#65)
- use securefix-server release workflow reusables (#62)
- update dependency jdx/mise to v2026.8.3 (#60)
- update dependency jdx/mise to v2026.8.2 (#59)
- update dependency jdx/mise to v2026.8.1 (#58)
- update dependency jdx/mise to v2026.8.0 (#57)
- update jdx/mise-action action to v4.2.4 (#56)
- update dependency jdx/mise to v2026.7.18 (#55)
- update dependency jdx/mise to v2026.7.17 (#54)
- update dependency aqua:suzuki-shunsuke/pinact to v4.1.1 (#53)
- update dependency jdx/mise to v2026.7.16 (#52)
- update dependency jdx/mise to v2026.7.15 (#51)
- update dependency jdx/mise to v2026.7.14 (#50)
- update dependency aqua:goreleaser/goreleaser to v2.17.1 (#49)
- lock file maintenance (#48)


### Refactor

- split provider types into one file each (#81)

## [0.1.1] - 2026-07-26


### Bug Fixes

- refuse SCIM member destroy and harden composite imports (#40)
- preserve base_url path and harden list/404 handling (#39)
- avoid attachment wipes and ambiguous tagged grant lookups (#38)
- invalidate cached token and retry once on 401 (#37)
- omit null member fields and reactivate archived members (#36)
- prevent write-only secret wipes on connection updates (#35)
- automerge non-major Renovate updates (#31)
- pass allowed_committers on Approve Request client (#32)
- floor Release PR base on shipped .release-version (#29)


### Documentation

- enrich Terraform Registry provider overview (#45)
- note tag protection and immutable releases (#41)


### Maintenance

- add identity and document data source UnitTests (#44)
- cover member PATCH omit and connection credential updates (#43)
- add identity resource mock UnitTests (#42)
- include dependabot in Approve Request actors (#34)
- disable dependency dashboard (#33)
- update jdx/mise-action action to v4.2.3 (#19)
- update dependency jdx/mise to v2026.7.13 (#13)
- update actions/checkout action to v7 (#20)
- update actions/create-github-app-token action to v3 (#24)
- auto-request approval for SecureFix-authored PRs (#28)

## [0.1.0] - 2026-07-24


### Bug Fixes

- harden Release Tag existing-tag check and allow dispatch (#26)
- use Securefix server repository name-only variable (#23)
- scan full history for first release version (#8)


### Documentation

- coverage matrix, contributing guide, and polish (#22)
- use main environment for securefix-server releases (#17)


### Features

- beta resources (tenants, deployment policies, source swap policies) (#21)
- document lifecycle resources and data sources (#18)
- connection resources (#15)
- workspace, file, and grant resources (#11)
- workspace, file, and grant resources (#10)
- identity and access resources (#9)
- provider skeleton, Sigma API client, and whoami data source (#3)


### Maintenance

- release pipeline (release PR, tagging, server-side goreleaser) (#7)
- add lint, test, and approval workflows (#2)
- repository foundation (#1)


### Miscellaneous

- Add initial README for terraform-provider-sigma


