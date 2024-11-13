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

# # all groups
# data "storagegrid_groups" "fetch_groups" {}

# output "fetch_groups" {
#   value = data.storagegrid_groups.fetch_groups
# }

# # by id
# data "storagegrid_group" "group_id" {
#   id = "aec838fe-523f-bd43-a4df-xxxxxx"
# }

# output "group_id" {
#   value = data.storagegrid_group.group_id
# }

# # by uniqueName - local
# data "storagegrid_group" "group_local_name" {
#   unique_name = "group/gitlab-s3"
# }

# output "group_local_name" {
#   value = data.storagegrid_group.group_local_name
# }

# # by federated name
# data "storagegrid_group" "group_fed_name" {
#   unique_name = "federated-group/xxxxx-xxxx"
# }

# output "group_fed_name" {
#   value = data.storagegrid_group.group_fed_name
# }

######
# Users
######
# data "storagegrid_users" "fetch_users" {}

# output "fetch_users" {
#   value = data.storagegrid_users.fetch_users
# }


# by id
# data "storagegrid_user" "user_id" {
#   id = "a74b96b2-4d44-8c4f-8bdb-xxxxxx"
# }

# output "user_id" {
#   value = data.storagegrid_user.user_id
# }

# # by uniqueName - local
# data "storagegrid_user" "user_local_name" {
#   unique_name = "user/gitlab-xxxxxx"
# }

# output "user_local_name" {
#   value = data.storagegrid_user.user_local_name
# }

# # by federated name
# data "storagegrid_user" "user_fed_name" {
#   unique_name = "federated-user/xxxxxxxxx"
# }

# output "user_fed_name" {
#   value = data.storagegrid_user.user_fed_name
# }