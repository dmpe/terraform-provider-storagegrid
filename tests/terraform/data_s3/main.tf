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
  unique_name = "user/test_creating_s3_keys"
  full_name   = "Test User for creating S3"
  disable     = "false"
  member_of = [
    "a9dd4848-a863-4716-82eb-d0939a6d643b"
  ]
}

data "storagegrid_user" "prevuser" {
  unique_name = "user/test_creating_s3_keys"
}

# all s3 access keys for specific user
data "storagegrid_s3_user_id_all_keys" "fetch_user_id_s3_access_keys" {
  user_uuid = data.storagegrid_user.prevuser.id
}

output "fetch_user_id_s3_access_keys" {
  value = data.storagegrid_s3_user_id_all_keys.fetch_user_id_s3_access_keys
}

output "fetch_user_id_s3_onespecific_key" {
  value = data.storagegrid_s3_user_id_all_keys.fetch_user_id_s3_access_keys.data[0].display_name
}

# a specific s3 access key for specific user
data "storagegrid_s3_user_id_access_key" "fetch_user_id_s3_specific_access_key" {
  user_uuid  = "940e50a9-bb1f-48d8-a371-c64bc7b0c1db"
  access_key = "xxxxxxxx" # data.storagegrid_s3_user_id_all_keys.fetch_user_id_s3_access_keys.data[0].display_name
}

output "fetch_user_id_s3_specific_access_key" {
  value = data.storagegrid_s3_user_id_access_key.fetch_user_id_s3_specific_access_key
}

