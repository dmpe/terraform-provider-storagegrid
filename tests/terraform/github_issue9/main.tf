# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

terraform {
  required_providers {
    storagegrid = {
      source = "github.com/dmpe/storagegrid"
    }
  }
}

provider "storagegrid" {
  address  = "https://grid.firm.com:9443/api/v3"
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_iid
  insecure = true
}

# Create multiple groups dynamically
resource "storagegrid_groups" "groups" {
  for_each       = { for g in var.groups : g.unique_name => g }
  unique_name    = each.value.unique_name
  display_name   = each.value.display_name
  management_read_only = each.value.management_read_only

  policies = {
    management = each.value.management_policies
    s3 = each.value.s3
  }
}

# Create multiple users dynamically
resource "storagegrid_users" "users" {
  for_each   = { for u in var.users : u.unique_name => u }
  unique_name = each.value.unique_name
  full_name   = each.value.full_name
  disable     = each.value.disable
  member_of   = [for group_name in each.value.member_of : storagegrid_groups.groups[group_name].id]
}

# Create multiple user's s3 access keys dynamically
resource "storagegrid_s3_access_key" "user_keys" {
    for_each    = { for u in var.users : u.unique_name => u if u.create_key == true }
    user_uuid   = storagegrid_users.users[each.value.unique_name].id
    expires     = each.value.key_expiry
}
