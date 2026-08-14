resource "sigma_user_attribute_tenant_assignment" "example" {
  user_attribute_id = "user-attribute-id"
  tenant_id         = "tenant-organization-id"
  value             = "tenant-connection"
}
