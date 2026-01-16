# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

provider "storagegrid" {
  address  = var.grid_url
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_iid
  insecure = true
}

resource "storagegrid_users" "new-local-user" {
  unique_name = "user/my_new_test_user_tf_stroragegrid_provider"
  full_name   = "My StorageGrid TF Provider plugin"
  disable     = "false"
  member_of   = var.group_memberships
}
