resource "sigma_user_attribute" "example" {
  name          = "Region"
  description   = "Data access region"
  default_value = "global"
}
