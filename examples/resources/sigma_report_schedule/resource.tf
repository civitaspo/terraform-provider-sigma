resource "sigma_report_schedule" "example" {
  report_id = "report-id"
  config_json = jsonencode({
    target = [{
      type      = "email"
      recipient = "analytics@example.com"
    }]
    schedule = {
      cronSpec = "0 9 * * 1"
    }
    configV2 = {
      title = "Weekly report export"
    }
  })
}
