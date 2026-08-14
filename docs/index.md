---
page_title: "Provider: Sigma"
description: |-
  The Sigma provider manages Sigma Computing resources through the Sigma REST API.
---

# Sigma Provider

The Sigma provider manages [Sigma Computing](https://www.sigmacomputing.com/)
organization resources through Terraform. Use it to configure members, teams,
workspaces, connections, grants, documents, schedules, and related Sigma
objects as code.

Provider address: `registry.terraform.io/civitaspo/sigma`

## Authentication

Configure API client credentials with the `client_id` and `client_secret`
arguments, or with environment variables:

- `SIGMA_CLIENT_ID`
- `SIGMA_CLIENT_SECRET`
- `SIGMA_BASE_URL` (required when `base_url` is unset)

Always set `base_url` (or `SIGMA_BASE_URL`) to the API host for the cloud
where your Sigma organization is hosted.

```terraform
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

## Resource and data source coverage

### Identity and access

| Resources | Data sources |
|-----------|--------------|
| `sigma_member` | `sigma_member`, `sigma_members` |
| `sigma_team` | `sigma_team`, `sigma_teams` |
| `sigma_team_member` | |
| `sigma_account_type` | `sigma_account_types` |
| `sigma_user_attribute` | `sigma_user_attributes` |
| `sigma_user_attribute_team_assignment` | |
| `sigma_user_attribute_user_assignment` | |

### Workspaces, files, and grants

| Resources | Data sources |
|-----------|--------------|
| `sigma_workspace` | `sigma_workspace`, `sigma_workspaces` |
| `sigma_folder` | `sigma_files` |
| `sigma_workspace_grant` | |
| `sigma_workbook_grant`, `sigma_report_grant` | |

### Connections

| Resources | Data sources |
|-----------|--------------|
| `sigma_connection` | `sigma_connection`, `sigma_connections` |
| `sigma_connection_grant`, `sigma_connection_path_grant` | `sigma_connection_paths` |
| `sigma_api_connector`, `sigma_api_credential` | |

Warehouse credentials use write-only attributes (`credentials_wo` +
`credentials_wo_version`). Because Sigma's connection update replaces warehouse
details entirely, any update after credentials were managed requires bumping
`credentials_wo_version` and resupplying `credentials_wo`. Connection tests
after create/update are warnings, not hard errors.

### Documents and schedules

| Resources | Data sources |
|-----------|--------------|
| `sigma_tag` | `sigma_tags` |
| `sigma_workbook_schedule`, `sigma_report_schedule` | `sigma_workbook`, `sigma_workbooks` |
| `sigma_workbook_embed` | `sigma_report`, `sigma_reports` |
| `sigma_translation` | `sigma_data_model`, `sigma_data_models` |
| | `sigma_dataset`, `sigma_datasets` (deprecated) |
| | `sigma_templates` |
| | `sigma_whoami` |

### Beta

These use Sigma Beta APIs and may change without notice.

| Resources | Data sources |
|-----------|--------------|
| `sigma_tenant` | `sigma_tenants` |
| `sigma_tenant_deployment_capability` | |
| `sigma_user_attribute_tenant_assignment` | |
| `sigma_deployment_policy` | `sigma_deployment_policies` |
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

## Requirements

Terraform >= 1.0. Write-only attributes (for example connection credentials)
require Terraform >= 1.11.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `base_url` (String) Sigma API base URL. May also be set with SIGMA_BASE_URL.
- `client_id` (String, Sensitive) Sigma API client ID. May also be set with SIGMA_CLIENT_ID.
- `client_secret` (String, Sensitive) Sigma API client secret. May also be set with SIGMA_CLIENT_SECRET.
