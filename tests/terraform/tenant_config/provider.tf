# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

terraform {
  required_providers {
    storagegrid = { source = "dmpe/storagegrid" }
  }
}

provider "storagegrid" {
  address  = var.grid_url
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_id
  insecure = true
}
