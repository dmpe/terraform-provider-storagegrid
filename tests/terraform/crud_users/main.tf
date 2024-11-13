# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = ">= 4.4.0"
    }
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

resource "storagegrid_users" "new-local-user" {
  unique_name = "user/my_new_test_user_tf_stroragegrid_provider"
  full_name   = "My StorageGrid TF Provider plugin"
  disable     = "false"
  member_of = [
    "a9dd4848-a863-4716-82eb-d0939a6d643b"
  ]
}
