# This will fetch (local or federated) group by specific ID
data "storagegrid_group" "group_id" {
  id = "aec838fe-523f-bd43-a4df-xxxxxx"
}

output "group_id" {
  value = data.storagegrid_group.group_id
}

# This will fetch local group by specific name (group/ prefix is mandatory)
data "storagegrid_group" "group_local_name" {
  unique_name = "group/gitlab-xxxxxx"
}

output "group_local_name" {
  value = data.storagegrid_group.group_local_name
}

# This will fetch federated group by specific name (federated-group/ prefix is mandatory)
data "storagegrid_group" "group_fed_name" {
  unique_name = "federated-group/xxxxxxxxx"
}

output "group_fed_name" {
  value = data.storagegrid_group.group_fed_name
}
