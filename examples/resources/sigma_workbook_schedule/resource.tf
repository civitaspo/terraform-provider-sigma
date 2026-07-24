resource "sigma_workbook_schedule" "example" {
  workbook_id = "workbook-id"
  config_json = jsonencode({
    target = [{
      type      = "email"
      recipient = "analytics@example.com"
    }]
    schedule = {
      cronSpec = "0 9 * * 1"
    }
    configV2 = {
      title = "Weekly workbook export"
    }
  })
}
