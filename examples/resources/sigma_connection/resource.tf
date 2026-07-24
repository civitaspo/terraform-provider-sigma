resource "sigma_connection" "example" {
  name = "analytics-postgres"

  details_json = jsonencode({
    type     = "postgres"
    host     = "database.example.com"
    database = "analytics"
    user     = "sigma"
    port     = 5432
    useTls   = true
  })

  credentials_wo = jsonencode({
    password = var.connection_password
  })

  credentials_wo_version = 1
}
