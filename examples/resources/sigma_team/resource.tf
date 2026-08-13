resource "sigma_team" "example" {
  name               = "Analytics"
  description        = "Analytics team"
  visibility         = "private"
  create_team_folder = true
}
