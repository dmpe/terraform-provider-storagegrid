# This will fetch (local or federated) user by specific ID
data "storagegrid_user" "user_id" {
  id = "a74b96b2-4d44-8c4f-8bdb-xxxxxx"
}

output "user_id" {
  value = data.storagegrid_user.user_id
}

# This will fetch local user by specific name (user/ prefix is mandatory)
data "storagegrid_user" "user_local_name" {
  unique_name = "user/gitlab-xxxxxx"
}

output "user_local_name" {
  value = data.storagegrid_user.user_local_name
}

# This will fetch federated user by specific name (federated-user/ prefix is mandatory)
data "storagegrid_user" "user_fed_name" {
  unique_name = "federated-user/xxxxxxxxx"
}

output "user_fed_name" {
  value = data.storagegrid_user.user_fed_name
}
