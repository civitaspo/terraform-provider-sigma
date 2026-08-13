resource "sigma_folder" "example" {
  name        = "Terraform Managed"
  parent_id   = sigma_workspace.example.id
  description = "Managed by Terraform"
}
