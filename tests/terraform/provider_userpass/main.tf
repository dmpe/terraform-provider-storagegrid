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

data "storagegrid_groups" "fetch_groups" {}

output "fetch_groups" {
  value = data.storagegrid_groups.fetch_groups
}
