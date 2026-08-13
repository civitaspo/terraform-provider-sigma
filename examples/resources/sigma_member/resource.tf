resource "sigma_member" "example" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  user_kind   = "internal"
  send_invite = true

  add_to_teams = [
    {
      team_id       = "team-1"
      is_team_admin = false
    }
  ]

  new_owner_id = "member-admin"
}
