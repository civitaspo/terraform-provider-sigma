# terraform-provider-sigma

[![CI](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/ci.yml/badge.svg)](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/ci.yml)
[![Release](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/release-tag.yml/badge.svg)](https://github.com/civitaspo/terraform-provider-sigma/actions/workflows/release-tag.yml)
[![Terraform Registry](https://img.shields.io/badge/registry-civitaspo%2Fsigma-purple.svg)](https://registry.terraform.io/providers/civitaspo/sigma/latest)

Terraform provider for [Sigma Computing](https://www.sigmacomputing.com/). Manage members, teams, workspaces, connections, grants, and related Sigma resources with Terraform.

> Status: under active construction. Early releases may only expose a subset of resources.

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
| AWS | `https://aws-api.sigmacomputing.com` |
| Azure | `https://api.sigmacomputing.com` (confirm for your tenant) |
| GCP | `https://api.sigmacomputing.com` (confirm for your tenant) |
| Other / custom | Use the API host shown in your Sigma admin / API docs |

Always set `base_url` explicitly for the cloud where your organization is hosted.

## Development quickstart

```bash
mise install --locked
mise run lint
mise run test
mise run build
```

See [AGENTS.md](AGENTS.md) for contribution conventions and [docs/securefix.md](docs/securefix.md) for automation details.

## License

[MIT](LICENSE) © civitaspo
