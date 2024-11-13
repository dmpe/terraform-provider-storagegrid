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

resource "storagegrid_users" "new-local-user-with-keys" {
  unique_name = "user/test_creating_s3_access_secret_keys"
  full_name   = "Test User for creating S3 keys"
  disable     = "false"
  member_of = [
    "a9dd4848-a863-4716-82eb-d0939a6d643b"
  ]
}

data "storagegrid_user" "prevuser-with-keys" {
  unique_name = "user/test_creating_s3_access_secret_keys"
}

resource "storagegrid_s3_access_key" "new-access-secret-key" {
  user_uuid = data.storagegrid_user.prevuser-with-keys.id
  #   example: 2028-09-04T00:00:00.000Z
  expires = "0"
}
