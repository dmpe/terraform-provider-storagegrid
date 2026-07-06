# Copyright github.com/dmpe 2024, 2026
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
