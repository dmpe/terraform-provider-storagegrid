# Copy this file to an empty directory as main.tf, or run it from this directory.
# It creates real StorageGRID tenant objects. Use a disposable tenant if possible.

terraform {
  required_providers {
    storagegrid = {
      source = "dmpe/storagegrid"
    }
  }
}

variable "grid_url" {
  description = "StorageGRID tenant API URL, for example https://grid.example.com:9443/api/v3"
  type        = string
}

variable "grid_username" {
  description = "StorageGRID tenant username"
  type        = string
}

variable "grid_password" {
  description = "StorageGRID tenant password"
  type        = string
  sensitive   = true
}

variable "grid_tenant_id" {
  description = "StorageGRID tenant/account ID"
  type        = string
}

variable "test_suffix" {
  description = "Unique lowercase suffix for this test run, for example your initials plus date: ll-20260508"
  type        = string
}

provider "storagegrid" {
  address  = var.grid_url
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_id
  insecure = true
}

locals {
  bucket_name = "tf-storagegrid-smoke-${var.test_suffix}"
  group_name  = "group/tf-storagegrid-smoke-${var.test_suffix}"
  user_name   = "user/tf-storagegrid-smoke-${var.test_suffix}"
}

resource "storagegrid_groups" "smoke" {
  unique_name          = local.group_name
  display_name         = "Terraform provider smoke test ${var.test_suffix}"
  management_read_only = false

  policies = {
    management = {
      manage_all_containers        = false
      manage_endpoints             = false
      manage_own_container_objects = false
      manage_own_s3_credentials    = true
      root_access                  = false
      view_all_containers          = false
    }

    s3 = {
      statement = [
        {
          sid      = "DenyDeleteOutsideSmokeTest"
          effect   = "Deny"
          action   = ["s3:DeleteObject"]
          resource = ["arn:aws:s3:::not-the-smoke-test-bucket/*"]
        }
      ]
    }
  }
}

resource "storagegrid_users" "smoke" {
  unique_name = local.user_name
  full_name   = "Terraform provider smoke test ${var.test_suffix}"
  disable     = false
  member_of   = [storagegrid_groups.smoke.id]
}

resource "storagegrid_bucket" "smoke" {
  name = local.bucket_name
}

resource "storagegrid_bucket_versioning" "smoke" {
  bucket_name = storagegrid_bucket.smoke.name
  status      = "Enabled"
}

resource "storagegrid_bucket_quota" "smoke" {
  bucket_name  = storagegrid_bucket.smoke.name
  object_bytes = 1073741824
}

resource "storagegrid_bucket_policy" "smoke" {
  bucket_name = storagegrid_bucket.smoke.name

  policy = {
    statement = [
      {
        sid    = "DenyDeleteForEveryone"
        effect = "Deny"
        action = ["s3:DeleteObject"]
        resource = [
          "arn:aws:s3:::${storagegrid_bucket.smoke.name}",
          "arn:aws:s3:::${storagegrid_bucket.smoke.name}/*"
        ]
        principal = {
          type = "*"
        }
      }
    ]
  }
}

resource "storagegrid_s3_access_key" "smoke" {
  user_uuid = storagegrid_users.smoke.id
  expires   = "0"
}

data "storagegrid_tenant_config" "current" {}

data "storagegrid_bucket" "smoke" {
  name = storagegrid_bucket.smoke.name
}

data "storagegrid_bucket_versioning" "smoke" {
  bucket_name = storagegrid_bucket.smoke.name
}

data "storagegrid_bucket_quota" "smoke" {
  bucket_name = storagegrid_bucket.smoke.name
}

data "storagegrid_group" "smoke" {
  unique_name = storagegrid_groups.smoke.unique_name
}

data "storagegrid_user" "smoke" {
  unique_name = storagegrid_users.smoke.unique_name
}

output "created_bucket" {
  value = data.storagegrid_bucket.smoke.name
}

output "bucket_versioning_status" {
  value = data.storagegrid_bucket_versioning.smoke.status
}

output "bucket_quota_bytes" {
  value = data.storagegrid_bucket_quota.smoke.object_bytes
}

output "created_group_id" {
  value = data.storagegrid_group.smoke.id
}

output "created_user_id" {
  value = data.storagegrid_user.smoke.id
}

output "s3_access_key" {
  value     = storagegrid_s3_access_key.smoke.access_key
  sensitive = true
}

output "s3_secret_access_key" {
  value     = storagegrid_s3_access_key.smoke.secret_access_key
  sensitive = true
}
