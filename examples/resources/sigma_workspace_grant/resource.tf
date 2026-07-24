resource "sigma_workspace_grant" "example" {
  inode_id   = sigma_workspace.example.id
  team_id    = "team-id"
  permission = "organize"
}
