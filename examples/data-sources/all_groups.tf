data "storagegrid_groups" "fetch_groups" {}

output "fetch_groups" {
  value = data.storagegrid_groups.fetch_groups
}
