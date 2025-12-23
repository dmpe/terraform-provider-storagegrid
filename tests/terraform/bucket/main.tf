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

resource "storagegrid_bucket" "example_compliance_object_lock" {
  name   = "${var.example_bucket.name}-compliance-object-lock"
  region = var.example_bucket.region

  object_lock_configuration {
    mode = "compliance"
    days = 30
  }
}

resource "storagegrid_bucket" "example_governance_object_lock" {
  name   = "${var.example_bucket.name}-governance-object-lock"
  region = var.example_bucket.region

  object_lock_configuration {
    mode = "governance"
    days = 30
  }
}

// Versioning is enabled by default if an `object_lock_configuration` is set on the bucket.
// Creating a `storagegrid_bucket_versioning` resource with status "Enabled" is possible nonetheless.
resource "storagegrid_bucket_versioning" "example_governance_object_lock" {
  bucket_name = storagegrid_bucket.example_governance_object_lock.name
  status      = "Enabled"
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

data "storagegrid_bucket_versioning" "read" {
  bucket_name = var.read_bucket_name
}

output "read_bucket_name" {
  value = data.storagegrid_bucket.read.name
}

output "read_bucket_region" {
  value = data.storagegrid_bucket.read.region
}

output "read_bucket_versioning_status" {
  value = data.storagegrid_bucket_versioning.read.status
}

# Scenario: Change bucket to which a `bucket_versioning` resource is attached.
# 1. After initial apply, un-comment the following resources.
# 3. Run `terraform apply`.
#
# resource "storagegrid_bucket" "change_version_bucket_name_target" {
#   name = "${var.import_bucket.name}-versioning"
# }
#
# resource "storagegrid_bucket_versioning" "enabled" {
#   bucket_name = storagegrid_bucket.change_version_bucket_name_target.name
#
#   status = "Enabled"
# }

# Szenario: Change bucket to which a `bucket_versioning` resource is attached.
# 2. Comment out the following resource `storagegrid_bucket_versioning.enabled`
resource "storagegrid_bucket_versioning" "enabled" {
  bucket_name = storagegrid_bucket.example.name

  status = "Enabled"
}

resource "storagegrid_bucket_versioning" "suspended" {
  bucket_name = storagegrid_bucket.example_default_region.name

  status = "Suspended"
}

// Specifying a `bucket_versioning` resource with status "Disabled" is allowed here since we actually import the
// resource.
// Otherwise, an API error would be generated.
resource "storagegrid_bucket_versioning" "disabled_import" {
  bucket_name = var.import_bucket.name

  status = "Disabled"
}

import {
  id = storagegrid_bucket.imported.name
  to = storagegrid_bucket_versioning.disabled_import
}

resource "storagegrid_bucket" "quota_enabled" {
  name = "${var.example_bucket.name}-quota-enabled"
}

resource "storagegrid_bucket_quota" "quota_enabled" {
  bucket_name = storagegrid_bucket.quota_enabled.name

  object_bytes = 10000000000
}

data "storagegrid_bucket_quota" "read_quota" {
  bucket_name = var.read_bucket_name
}

output "read_bucket_quota" {
  value = data.storagegrid_bucket_quota.read_quota.object_bytes
}

data "storagegrid_user" "root_user" {
  unique_name = "root"
}

resource "storagegrid_bucket_policy" "example" {
  bucket_name = storagegrid_bucket.example_default_region.name

  policy = {
    statement = [{
      sid    = "example-sid"
      effect = "Allow"
      action = ["s3:ListBucket"]
      resource = [
        "arn:aws:s3:::${storagegrid_bucket.example_default_region.name}",
        "arn:aws:s3:::${storagegrid_bucket.example_default_region.name}/*"
      ]
      principal = {
        type = "AWS"
        identifiers = [
          "arn:aws:iam::${var.grid_tenant_iid}:${data.storagegrid_user.root_user.unique_name}"
        ]
      }
    }]
  }
}
