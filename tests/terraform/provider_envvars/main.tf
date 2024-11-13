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
  insecure = true
}

data "storagegrid_groups" "fetch_groups" {}

output "fetch_groups" {
  value = data.storagegrid_groups.fetch_groups
}
