resource "sigma_connection_grant" "example" {
  connection_id = sigma_connection.example.id
  team_id       = "team-id"
  permission    = "usage"
}
