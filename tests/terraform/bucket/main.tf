# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

provider "storagegrid" {
  address  = var.grid_url
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_iid
  insecure = true
}

resource "storagegrid_bucket" "example" {
  name   = var.example_bucket.name
  region = var.example_bucket.region
}

# uses default region
resource "storagegrid_bucket" "example_default_region" {
  name = "${var.example_bucket.name}-default-region"
}

import {
  id = var.import_bucket.name
  to = storagegrid_bucket.imported
}

resource "storagegrid_bucket" "imported" {
  name   = var.import_bucket.name
  region = var.import_bucket.region
}

data "storagegrid_bucket" "read" {
  name = var.read_bucket_name
}

output "read_bucket_name" {
  value = data.storagegrid_bucket.read.name
}

output "read_bucket_region" {
  value = data.storagegrid_bucket.read.region
}
