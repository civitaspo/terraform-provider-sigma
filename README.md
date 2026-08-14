# terraform-provider-sigma

[![CI](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/pull_request.yml/badge.svg)](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/pull_request.yml)
[![Release](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/release-tag.yml/badge.svg)](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/release-tag.yml)
[![Terraform Registry](https://img.shields.io/badge/registry-civitaspo%2Fsigma-purple.svg)](https://registry.terraform.io/providers/civitaspo/sigma/latest)

Terraform provider for [Sigma Computing](https://www.sigmacomputing.com/). Manage members, teams, workspaces, connections, grants, documents, and related Sigma resources with Terraform.

Provider address: `registry.terraform.io/civitaspo/sigma`

## Requirements

| Name | Version |
|------|---------|
| [Terraform](https://www.terraform.io/downloads.html) | >= 1.0 (write-only attributes require >= 1.11) |
| [Go](https://go.dev/dl/) (development) | 1.26.x |
| [mise](https://mise.jdx.dev/) (development) | latest |

## Provider configuration

```hcl
terraform {
  required_providers {
    sigma = {
      source  = "civitaspo/sigma"
      version = "~> 0.1"
    }
  }
}

provider "sigma" {
  # Required (or set SIGMA_BASE_URL)
  base_url = "https://aws-api.sigmacomputing.com"

  # Optional when set via environment variables
  # client_id     = var.sigma_client_id
  # client_secret = var.sigma_client_secret
}
```

Environment variable fallbacks:

- `SIGMA_BASE_URL` (required if `base_url` is unset)
- `SIGMA_CLIENT_ID`
- `SIGMA_CLIENT_SECRET`

### Per-cloud `base_url` values

| Cloud | Base URL |
|-------|----------|
| AWS US (West) | `https://aws-api.sigmacomputing.com` |
| AWS US (East) | `https://api.us-a.aws.sigmacomputing.com` |
| AWS Canada | `https://api.ca.aws.sigmacomputing.com` |
| AWS Europe | `https://api.eu.aws.sigmacomputing.com` |
| AWS UK | `https://api.uk.aws.sigmacomputing.com` |
| AWS Australia / APAC | `https://api.au.aws.sigmacomputing.com` |
| GCP (US) | `https://api.sigmacomputing.com` |
| GCP (KSA) | `https://api.sa.gcp.sigmacomputing.com` |
| Azure US | `https://api.us.azure.sigmacomputing.com` |
| Azure Europe | `https://api.eu.azure.sigmacomputing.com` |
| Azure Canada | `https://api.ca.azure.sigmacomputing.com` |
| Azure UK | `https://api.uk.azure.sigmacomputing.com` |
| Azure Australia | `https://api.au.azure.sigmacomputing.com` |

Always set `base_url` explicitly for the cloud where your organization is hosted.

## Resource and data source coverage

### Identity and access

| Resources | Data sources |
|-----------|--------------|
| `sigma_member` | `sigma_member`, `sigma_members` |
| `sigma_team` | `sigma_team`, `sigma_teams` |
| `sigma_team_member` | |
| `sigma_account_type` | `sigma_account_types` |
| `sigma_user_attribute` | `sigma_user_attribute`, `sigma_user_attributes` |
| `sigma_user_attribute_team_assignment` | |
| `sigma_user_attribute_user_assignment` | |

### Workspaces, files, and grants

| Resources | Data sources |
|-----------|--------------|
| `sigma_workspace` | `sigma_workspace`, `sigma_workspaces` |
| `sigma_folder` | `sigma_file`, `sigma_files` |
| `sigma_workspace_grant` | |
| `sigma_workbook_grant`, `sigma_report_grant` | |

### Connections

| Resources | Data sources |
|-----------|--------------|
| `sigma_connection` | `sigma_connection`, `sigma_connections` |
| `sigma_connection_grant`, `sigma_connection_path_grant` | `sigma_connection_path`, `sigma_connection_paths` |
| `sigma_api_connector`, `sigma_api_credential` | |

Warehouse credentials use write-only attributes (`credentials_wo` + `credentials_wo_version`). Because Sigma's connection update replaces warehouse details entirely, any update after credentials were managed requires bumping `credentials_wo_version` and resupplying `credentials_wo`. Connection restore is not a Terraform attribute. Connection tests after create/update are warnings, not hard errors.

### Documents and schedules

| Resources | Data sources |
|-----------|--------------|
| `sigma_tag` | `sigma_tags` |
| `sigma_workbook_schedule`, `sigma_report_schedule` | `sigma_workbook`, `sigma_workbooks` |
| `sigma_workbook_embed` | `sigma_report`, `sigma_reports` |
| `sigma_translation` | `sigma_data_model`, `sigma_data_models` |
| | `sigma_dataset`, `sigma_datasets` (deprecated) |
| | `sigma_template`, `sigma_templates` |
| | `sigma_whoami` |

### Beta

These use Sigma Beta APIs and may change without notice.

| Resources | Data sources |
|-----------|--------------|
| `sigma_tenant` | `sigma_tenant`, `sigma_tenants` |
| `sigma_tenant_deployment_capability` | |
| `sigma_user_attribute_tenant_assignment` | |
| `sigma_deployment_policy` | `sigma_deployment_policy`, `sigma_deployment_policies` |
| `sigma_deployment_policy_document` | |
| `sigma_deployment_policy_tenant` | |
| `sigma_source_swap_policy` | |

## Out of scope

The following Sigma capabilities are intentionally not managed by this provider:

| Area | Rationale |
|------|-----------|
| Query execution / exports | Operational, not durable infrastructure |
| Lineage | Read-heavy analytics surface; poor Terraform fit |
| Materialization runs | Ephemeral job orchestration |
| Favorites | Per-user preferences |
| SAML certificate management | Excluded by design |
| Embed URL generation / JWT signing | Host-app concern; prefer JWT-signed URLs |
| Workbook duplication / one-shot source swap actions | Imperative actions, not long-lived resources |
| Organization API client keys (`/v2/credentials`) | Distinct from third-party `sigma_api_credential` |
| Dedicated workbook/report resources | Short lifecycle and high AI change frequency are a poor Terraform fit. Data sources and specialized grants remain; folders use `sigma_folder` and workspaces use `sigma_workspace` |
| Generic inode grants (`sigma_grant`) | Overlapping ownership; use `sigma_workspace_grant`, `sigma_workbook_grant`, or `sigma_report_grant` |
| Aggregate team membership (`sigma_team_members`) | Overlapping ownership; use singular `sigma_team_member` |
| Applying version tags to documents | Short lifecycle and high AI change frequency are a poor Terraform fit. Tag definitions via `sigma_tag` remain |
| Data model spec / as-code content | Short lifecycle and high AI change frequency are a poor Terraform fit |

## Acceptance tests

Unit tests use a mock Sigma HTTP server and fail if handwritten `internal/provider` or `internal/sigma` coverage is below 80%:

```bash
mise run test
```

Acceptance tests require real credentials and `TF_ACC=1`:

```bash
export TF_ACC=1
export SIGMA_BASE_URL="https://aws-api.sigmacomputing.com"
export SIGMA_CLIENT_ID="..."
export SIGMA_CLIENT_SECRET="..."
go test ./internal/provider/ -run TestAcc -count=1 -timeout 30m
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full contributor guidance.

## Development quickstart

```bash
mise install --locked
mise run lint
mise run test
mise run build
mise run docs
```

See [AGENTS.md](AGENTS.md) for coding-agent conventions, [docs/securefix.md](docs/securefix.md) for automation details, and [docs/releasing.md](docs/releasing.md) for the release pipeline.

## License

[MIT](LICENSE) © civitaspo
