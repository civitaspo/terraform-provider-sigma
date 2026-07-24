resource "sigma_grant" "example" {
  inode_id   = sigma_file.example.id
  member_id  = "member-id"
  permission = "view"
}
