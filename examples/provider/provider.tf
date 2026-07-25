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
