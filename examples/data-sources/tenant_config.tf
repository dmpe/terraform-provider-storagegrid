data "storagegrid_tenant_config" "current" {}

output "example_tenant_config" {
  value = data.storagegrid_tenant_config.current
}
