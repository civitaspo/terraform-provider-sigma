data "sigma_connection_path" "example" {
  connection_id = "connection-id"
  path          = ["DATABASE", "SCHEMA", "TABLE"]
}
