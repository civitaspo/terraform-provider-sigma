resource "sigma_member" "example" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  user_kind   = "internal"
  send_invite = true

  new_owner_id = "member-admin"
}
