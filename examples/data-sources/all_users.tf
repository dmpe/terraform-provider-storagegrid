data "storagegrid_users" "fetch_users" {}

output "fetch_users" {
  value = data.storagegrid_users.fetch_users
}
