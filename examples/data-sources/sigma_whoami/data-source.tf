data "sigma_whoami" "current" {}

output "sigma_user_id" {
  value = data.sigma_whoami.current.user_id
}
