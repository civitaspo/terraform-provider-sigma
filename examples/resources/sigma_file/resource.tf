resource "sigma_file" "example" {
  type        = "folder"
  name        = "Terraform Managed"
  parent_id   = sigma_workspace.example.id
  description = "Managed by Terraform"
}
