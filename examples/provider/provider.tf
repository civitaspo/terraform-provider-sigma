terraform {
  required_providers {
    sigma = {
      source = "civitaspo/sigma"
    }
  }
}

provider "sigma" {
  base_url = "https://api.sigmacomputing.com"
}
