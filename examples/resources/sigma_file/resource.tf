resource "sigma_file" "example" {
  type        = "folder"
  name        = "Terraform Managed"
  parent_id   = sigma_workspace.example.id
  description = "Managed by Terraform"
}

resource "sigma_file" "from_source" {
  type            = "workbook"
  name            = "Copied workbook"
  parent_id       = sigma_workspace.example.id
  source_inode_id = "workbook-source-id"
  source_version  = 1
}
