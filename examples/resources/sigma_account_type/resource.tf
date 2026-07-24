resource "sigma_account_type" "example" {
  name        = "Analyst"
  description = "Custom analyst permissions"
  permissions = ["view-worksheet"]
}
