resource "sigma_tenant_deployment_capabilities" "example" {
  tenant_id    = "source-tenant-id"
  capabilities = ["target-tenant-id"]
}
