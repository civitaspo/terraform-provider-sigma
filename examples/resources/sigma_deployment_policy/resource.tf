resource "sigma_deployment_policy" "example" {
  name                 = "Starter pack"
  version_tag_id       = "version-tag-id"
  source_swap_policies = ["source-swap-policy-id"]
  inode_ids            = ["workbook-or-datamodel-id"]
  tenant_ids           = ["tenant-organization-id"]
}
