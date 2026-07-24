data "sigma_connection_paths" "example" {
  connection_id = "connection-id"
}

data "sigma_connection_paths" "lookup" {
  connection_id = "connection-id"
  path          = ["DATABASE", "SCHEMA", "TABLE"]
}
