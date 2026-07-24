resource "sigma_api_connector" "example" {
  name        = "weather"
  description = "Fetch current weather"
  auth_id     = sigma_api_credential.example.id

  params_json = jsonencode({
    method      = "GET"
    url         = "https://api.example.com/weather"
    headers     = []
    pathParams  = []
    queryParams = []
  })
}
