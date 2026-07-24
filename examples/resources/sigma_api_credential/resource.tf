resource "sigma_api_credential" "example" {
  name        = "weather-api"
  description = "Credential for the weather API"
  allowlist   = ["api.example.com"]

  credential_wo = jsonencode({
    authMethod = "apiKey"
    apiKey = {
      key          = "X-API-Key"
      value        = var.weather_api_key
      isQueryParam = false
    }
  })

  credential_wo_version = 1
}
