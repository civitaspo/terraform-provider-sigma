resource "sigma_connection_path_grant" "example" {
  connection_path_id = "connection-path-id"
  team_id             = "team-id"
  permission          = "annotate"
}
